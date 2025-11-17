package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"prutya/go-api-template/internal/logger"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

var ErrEmailDomainNotAllowed = errors.New("email domain not allowed")

func (s *authenticationService) Register(ctx context.Context, email string, password string) error {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	// Check if the email domain is allowed
	if !s.isEmailDomainAllowed(email) {
		return ErrEmailDomainNotAllowed
	}

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
				logger.DebugContext(ctx, "user not found", "email", email)
				// User not found, continue
			} else {
				return err
			}
		}

		verificationExpiresAt := time.Now().UTC().Add(s.config.AuthenticationEmailVerificationCodeTTL)
		veriticationCooldownResetsAt := time.Now().UTC().Add(s.config.AuthenticationEmailVerificationCooldown)

		if user == nil {
			// Create a new user
			newUUID, err := generateUUID()
			if err != nil {
				return err
			}
			userID = newUUID

			passwordDigest, err := s.argon2GenerateHashFromPassword(password)
			if err != nil {
				return err
			}

			if err := userRepo.Create(
				ctx,
				userID,
				email,
				passwordDigest,
				verificationExpiresAt,
				veriticationCooldownResetsAt,
			); err != nil {
				// Handle unique constraint error
				if pgErr, isPgErr := err.(*pgconn.PgError); isPgErr && pgErr.Code == "23505" {
					logger.DebugContext(ctx, ErrUserAlreadyExists.Error(), "user_id", userID, "email", email)

					return ErrUserAlreadyExists
				}

				return err
			}
		} else {
			userID = user.ID

			// If the email is already registered, and is verified, do nothing
			if user.EmailVerifiedAt.Valid {
				logger.DebugContext(ctx, ErrEmailAlreadyVerified.Error(), "user_id", userID, "email", email)

				return ErrEmailAlreadyVerified
			}

			// Check cooldown
			if user.EmailVerificationCooldownResetsAt.Valid && user.EmailVerificationCooldownResetsAt.Time.After(time.Now().UTC()) {
				logger.DebugContext(ctx, ErrEmailVerificationCooldown.Error(), "user_id", userID, "email", email)

				return ErrEmailVerificationCooldown
			}

			// Reset OTP hash and attempts, update cooldown and OTP expiration time
			if err := userRepo.StartEmailVerification(
				ctx,
				userID,
				verificationExpiresAt,
				veriticationCooldownResetsAt,
			); err != nil {
				return err
			}
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
