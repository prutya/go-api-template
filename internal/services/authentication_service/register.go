package authentication_service

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailDomainNotAllowed = errors.New("email domain not allowed")

// NOTE: I am not using transactions here, because it's just a single write
// operation
func (s *authenticationService) Register(ctx context.Context, email string, password string) error {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	// Normalize email

	email = normalizeEmail(email)

	// Check if the email domain is allowed
	if !s.isEmailDomainAllowed(email) {
		return ErrEmailDomainNotAllowed
	}

	// Check if the email is already registered

	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User not found, continue
		} else {
			return err
		}
	}

	if user != nil {
		// If the email is already registered, and is verified, do nothing
		if user.EmailVerifiedAt.Valid {
			return nil
		}

		// If the email is already registered, and is not verified, send a new
		// verification email
		if err := s.scheduleEmailVerification(ctx, user.ID); err != nil {
			return err
		}

		return nil
	}

	// Create a new user

	userID, err := generateUUID()
	if err != nil {
		return err
	}

	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(password), s.config.AuthenticationBcryptCost)
	if err != nil {
		return err
	}

	err = userRepo.Create(ctx, userID, email, string(passwordDigest))
	if err != nil {
		return err
	}

	// Schedule a verification email

	if err := s.scheduleEmailVerification(ctx, userID); err != nil {
		return err
	}

	return nil
}
