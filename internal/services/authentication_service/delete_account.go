package authentication_service

import (
	"context"
	"prutya/go-api-template/internal/argon2_utils"
)

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

	// Check if the password is correct
	passwordMatch, err := argon2_utils.Compare(password, user.PasswordDigest)
	if err != nil {
		return err
	}

	if !passwordMatch {
		return ErrInvalidCredentials
	}

	// Delete the user
	if err := userRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	return nil
}
