package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

var ErrEmailDomainNotAllowed = errors.New("email domain not allowed")

func (s *authenticationService) Register(ctx context.Context, email string, password string) error {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	// Check if the email domain is allowed
	if !s.isEmailDomainAllowed(email) {
		return ErrEmailDomainNotAllowed
	}

	var userID string

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := s.repoFactory.NewUserRepo(tx)

		user, err := userRepo.FindByEmailForUpdateNowait(ctx, email)
		if err != nil {
			// Handle postgres lock error
			if pgErr, isPgErr := err.(*pgconn.PgError); isPgErr && pgErr.Code == "55P03" {
				return ErrUserLocked
			}

			if errors.Is(err, sql.ErrNoRows) {
				// User not found, continue
			} else {
				return err
			}
		}

		if user != nil {
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
			if err := userRepo.ResetEmailVerification(
				ctx,
				userID,
				// TODO: Extract configuration variables
				time.Now().UTC().Add(15*time.Minute),
				time.Now().UTC().Add(1*time.Minute),
			); err != nil {
				return err
			}

			return nil
		}

		// Create a new user
		userID_tx, err := generateUUID()
		if err != nil {
			return err
		}
		userID = userID_tx

		// TODO: Extract configuration variables
		passwordDigest, err := argon2GenerateHashFromPassword(
			password,
			&argon2params{
				memory:      64 * 1024,
				iterations:  3,
				parallelism: 2,
				saltLength:  16,
				keyLength:   32,
			},
		)
		if err != nil {
			return err
		}

		err = userRepo.Create(ctx, userID, email, passwordDigest)
		if err != nil {
			// Handle unique constraint error
			if pgErr, isPgErr := err.(*pgconn.PgError); isPgErr && pgErr.Code == "23505" {
				return ErrUserAlreadyExists
			}

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
