package authentication_service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// NOTE: I am not using transactions here, because it's just a single write
// operation
func (s *authenticationService) DeleteAccount(
	ctx context.Context,
	accessTokenClaims *AccessTokenClaims,
	password string,
) error {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	userRepo := s.repoFactory.NewUserRepo(s.db)

	// Find the user by ID
	user, err := userRepo.FindByID(ctx, accessTokenClaims.UserID)
	if err != nil {
		return err
	}

	// Verify the password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}

		return err
	}

	// Delete the user
	if err := userRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	return nil
}
