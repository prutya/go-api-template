package authentication_service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/argon2"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/tasks"
)

func (s *authenticationService) scheduleEmailVerification(ctx context.Context, userID string) error {
	task, err := tasks.NewSendVerificationEmailTask(userID)

	if err != nil {
		return err
	}

	_, err = s.tasksClient.Enqueue(ctx, task)

	if err != nil {
		return err
	}

	return nil
}

func (s *authenticationService) isEmailDomainAllowed(email string) bool {
	domain := strings.Split(email, "@")[1]
	domain = strings.ToLower(domain)

	_, blocked := s.config.AuthenticationEmailBlocklist[domain]

	return !blocked
}

func findUserByID(ctx context.Context, userRepo repo.UserRepo, userID string) (*models.User, error) {
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.MustFromContext(ctx).WarnContext(ctx, "user not found", "user_id", userID, "error", err)

			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

// Ensures the function takes at least the specified minimum duration to
// execute. This is useful for preventing timing attacks by adding a delay to
// the function execution time.
func withMinimumAllowedFunctionDuration(ctx context.Context, minimumAllowedFunctionDuration time.Duration) func() {
	startTime := time.Now()

	return func() {
		duration := time.Since(startTime)
		timeLeft := minimumAllowedFunctionDuration - duration

		logger.MustDebugContext(ctx, "Function has returned", "real_duration", duration)

		if timeLeft > 0 {
			time.Sleep(timeLeft)
		}
	}
}

func generateUUID() (string, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

func generateRandomBytes(length uint32) ([]byte, error) {
	secret := make([]byte, length)

	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	return secret, nil
}

var errArgon2InvalidHash = errors.New("the encoded hash is not in the correct format")
var errArgon2IncompatibleVersion = errors.New("incompatible version of argon2")

type argon2params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func (s *authenticationService) argon2GenerateHashFromPassword(password string) (string, error) {
	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(s.config.AuthenticationArgon2SaltLength)
	if err != nil {
		return "", err
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		s.config.AuthenticationArgon2Iterations,
		s.config.AuthenticationArgon2Memory,
		s.config.AuthenticationArgon2Parallelism,
		s.config.AuthenticationArgon2KeyLength,
	)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.config.AuthenticationArgon2Memory,
		s.config.AuthenticationArgon2Iterations,
		s.config.AuthenticationArgon2Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func argon2ComparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := argon2DecodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func argon2DecodeHash(encodedHash string) (p *argon2params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errArgon2InvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errArgon2IncompatibleVersion
	}

	p = &argon2params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func generateOtp() (string, error) {
	max := big.NewInt(1000000) // 0 to 999999

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Format with leading zeros to ensure 6 digits
	return fmt.Sprintf("%06d", n), nil
}

func generateHmac(input []byte, secret []byte) ([]byte, error) {
	h := hmac.New(sha256.New, secret)

	if _, err := h.Write(input); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func checkHmac(input []byte, secret []byte, expected []byte) (bool, error) {
	inputHmac, err := generateHmac(input, secret)
	if err != nil {
		return false, err
	}

	return hmac.Equal(inputHmac, expected), nil
}
