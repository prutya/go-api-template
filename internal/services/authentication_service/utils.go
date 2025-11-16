package authentication_service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"

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

func generateSecret(length int) ([]byte, error) {
	secret := make([]byte, length)

	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	return secret, nil
}
