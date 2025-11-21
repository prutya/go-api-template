package authentication_service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"

	"prutya/go-api-template/internal/argon2_utils"
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

func (s *authenticationService) argon2GenerateHashFromPassword(password string) (string, error) {
	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(s.config.AuthenticationPasswordArgon2SaltLength)
	if err != nil {
		return "", err
	}

	return argon2_utils.CalculateAndEncode(
		[]byte(password),
		salt,
		&argon2_utils.Params{
			Memory:      s.config.AuthenticationPasswordArgon2Memory,
			Iterations:  s.config.AuthenticationPasswordArgon2Iterations,
			Parallelism: s.config.AuthenticationPasswordArgon2Parallelism,
			KeyLength:   s.config.AuthenticationPasswordArgon2KeyLength,
		},
	), nil
}

func (s *authenticationService) argon2GenerateHashFromOTP(otp string) (string, error) {
	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(s.config.AuthenticationOTPArgon2SaltLength)
	if err != nil {
		return "", err
	}

	return argon2_utils.CalculateAndEncode(
		[]byte(otp),
		salt,
		&argon2_utils.Params{
			Memory:      s.config.AuthenticationOTPArgon2Memory,
			Iterations:  s.config.AuthenticationOTPArgon2Iterations,
			Parallelism: s.config.AuthenticationOTPArgon2Parallelism,
			KeyLength:   s.config.AuthenticationOTPArgon2KeyLength,
		},
	), nil
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
