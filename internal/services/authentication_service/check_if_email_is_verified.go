package authentication_service

import (
	"context"
	"database/sql"
	"errors"
)

var ErrEmailNotVerified = errors.New("email not verified")

func (s *authenticationService) CheckIfEmailIsVerified(ctx context.Context, userID string) error {
	userRepo := s.repoFactory.NewUserRepo(s.db)

	// Find the user by ID
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}

		return err
	}

	if !user.EmailVerifiedAt.Valid {
		return ErrEmailNotVerified
	}

	return nil
}
