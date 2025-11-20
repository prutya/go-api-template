package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/tasks"
)

func (s *authenticationService) RequestPasswordReset(ctx context.Context, email string) error {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	logger := logger.MustFromContext(ctx)

	var userID string

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := s.repoFactory.NewUserRepo(tx)

		user, err := userRepo.FindByEmailForUpdateNowait(ctx, email)
		if err != nil {
			// Handle postgres lock error
			if pgErr, isPgErr := err.(*pgconn.PgError); isPgErr && pgErr.Code == "55P03" {
				logger.DebugContext(ctx, pgErr.Error(), "email", email)

				return ErrUserRecordLocked
			}

			if errors.Is(err, sql.ErrNoRows) {
				logger.DebugContext(ctx, ErrUserNotFound.Error(), "email", email)

				return ErrUserNotFound
			}

			return err
		}

		userID = user.ID

		// Check cooldown
		if user.PasswordResetCooldownResetsAt.Valid && user.PasswordResetCooldownResetsAt.Time.After(time.Now().UTC()) {
			logger.DebugContext(ctx, ErrPasswordResetCooldown.Error(), "user_id", userID, "email", email)

			return ErrPasswordResetCooldown
		}

		currentTime := time.Now().UTC()

		// Reset OTP hash and attempts, update cooldown and OTP expiration time
		if err := userRepo.StartPasswordReset(
			ctx,
			userID,
			currentTime.Add(s.config.AuthenticationPasswordResetCodeTTL),
			currentTime.Add(s.config.AuthenticationPasswordResetCooldown),
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Schedule a password reset email
	task, err := tasks.NewSendPasswordResetEmailTask(userID)
	if err != nil {
		return err
	}

	_, err = s.tasksClient.Enqueue(ctx, task)
	if err != nil {
		return err
	}

	return nil
}
