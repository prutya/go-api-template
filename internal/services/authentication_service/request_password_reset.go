package authentication_service

import (
	"context"
	"database/sql"
	"errors"

	"prutya/go-api-template/internal/tasks"
)

// NOTE: I am not using transactions here, because it's just a read operation
func (s *authenticationService) RequestPasswordReset(ctx context.Context, email string) error {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	userRepo := s.repoFactory.NewUserRepo(s.db)

	// Find the user by email
	user, err := userRepo.FindByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}

		return err
	}

	// Schedule a password reset email
	task, err := tasks.NewSendPasswordResetEmailTask(user.ID)
	if err != nil {
		return err
	}

	_, err = s.tasksClient.Enqueue(ctx, task)
	if err != nil {
		return err
	}

	return nil
}
