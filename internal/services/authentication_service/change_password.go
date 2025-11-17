package authentication_service

import (
	"context"
	"errors"
)

// NOTE: I am not using transactions here, because if the password is changed
// but there are still unterminated sessions, it's not a big deal, since we
// can still terminate them manually
func (s *authenticationService) ChangePassword(
	ctx context.Context,
	currentAccessTokenClaims *AccessTokenClaims,
	oldPassword string,
	newPassword string,
	terminateOtherSessions bool,
) error {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	return errors.New("not implemented")

	// userRepo := s.repoFactory.NewUserRepo(s.db)
	// sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	// // Find the user by ID
	// user, err := userRepo.FindByID(ctx, currentAccessTokenClaims.UserID)
	// if err != nil {
	// 	return err
	// }

	// // Find the current session
	// session, err := sessionRepo.FindByAccessTokenID(ctx, currentAccessTokenClaims.ID)
	// if err != nil {
	// 	return err
	// }

	// // Check if the old password is correct
	// if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(oldPassword)); err != nil {
	// 	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
	// 		return ErrInvalidCredentials
	// 	}

	// 	return err
	// }

	// // Hash the new password
	// newPasswordDigest, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.AuthenticationBcryptCost)
	// if err != nil {
	// 	return err
	// }

	// // Update the user's password
	// if err := userRepo.UpdatePasswordDigest(ctx, user.ID, string(newPasswordDigest)); err != nil {
	// 	return err
	// }

	// // Terminate other sessions if requested
	// if terminateOtherSessions {
	// 	if err := sessionRepo.TerminateAllSessionsExceptCurrentByUserID(ctx, user.ID, session.ID); err != nil {
	// 		return err
	// 	}
	// }

	// return nil
}
