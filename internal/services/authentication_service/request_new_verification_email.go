package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

func (s *authenticationService) RequestNewVerificationEmail(ctx context.Context, email string) error {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	var userID string

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := s.repoFactory.NewUserRepo(tx)

		user, err := userRepo.FindByEmailForUpdateNowait(ctx, email)
		if err != nil {
			// Handle postgres lock error
			if pgErr, isPgErr := err.(*pgconn.PgError); isPgErr && pgErr.Code == "55P03" {
				return ErrUserRecordLocked
			}

			if errors.Is(err, sql.ErrNoRows) {
				return ErrUserNotFound
			}

			return err
		}

		userID = user.ID

		// If the email is already registered, and is verified, do nothing
		if user.EmailVerifiedAt.Valid {
			return ErrEmailAlreadyVerified
		}

		// Check cooldown
		if user.EmailVerificationCooldownResetsAt.Valid && user.EmailVerificationCooldownResetsAt.Time.After(time.Now().UTC()) {
			return ErrEmailVerificationCooldown
		}

		// Reset OTP hash and attempts, update cooldown and OTP expiration time
		if err := userRepo.StartEmailVerification(
			ctx,
			userID,
			time.Now().UTC().Add(s.config.AuthenticationEmailVerificationCodeTTL),
			time.Now().UTC().Add(s.config.AuthenticationEmailVerificationCooldown),
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Schedule a verification email

	if err := s.scheduleEmailVerification(ctx, userID); err != nil {
		return err
	}

	return nil
}
