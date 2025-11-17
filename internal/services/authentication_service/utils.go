package authentication_service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
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
func withMinimumAllowedFunctionDuration(minimumAllowedFunctionDuration time.Duration) func() {
	startTime := time.Now()

	return func() {
		duration := time.Since(startTime)
		timeLeft := minimumAllowedFunctionDuration - duration

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

func generateRandomBytes(length int) ([]byte, error) {
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

func argon2GenerateHashFromPassword(
	password string,
	p *argon2params,
) (string, error) {
	// Generate a cryptographically secure random salt.
	// TODO: See if we can use uint32 in our config file
	salt, err := generateRandomBytes(int(p.saltLength))
	if err != nil {
		return "", err
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
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
