package authentication_service

import (
	"context"
)

func (s *authenticationService) ChangePassword(
	ctx context.Context,
	currentAccessTokenClaims *AccessTokenClaims,
	oldPassword string,
	newPassword string,
	terminateOtherSessions bool,
) error {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	userRepo := s.repoFactory.NewUserRepo(s.db)
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	// Find the user by ID
	user, err := userRepo.FindByID(ctx, currentAccessTokenClaims.UserID)
	if err != nil {
		return err
	}

	// Find the current session
	session, err := sessionRepo.FindByAccessTokenID(ctx, currentAccessTokenClaims.ID)
	if err != nil {
		return err
	}

	// Check if the password is correct
	passwordMatch, err := argon2ComparePasswordAndHash(oldPassword, user.PasswordDigest)
	if err != nil {
		return err
	}

	if !passwordMatch {
		return ErrInvalidCredentials
	}

	// Hash the new password
	newPasswordDigest, err := s.argon2GenerateHashFromPassword(newPassword)
	if err != nil {
		return err
	}

	// Update the user's password
	if err := userRepo.ChangePassword(ctx, user.ID, string(newPasswordDigest)); err != nil {
		return err
	}

	// Terminate other sessions if requested
	if terminateOtherSessions {
		if err := sessionRepo.TerminateAllSessionsExceptCurrentByUserID(ctx, user.ID, session.ID); err != nil {
			return err
		}
	}

	return nil
}
