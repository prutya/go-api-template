package authentication_service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
)

func (s *authenticationService) Login(
	ctx context.Context,
	email string,
	password string,
	userAgent string,
	ipAddress string,
) (*CreateTokensResult, error) {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	userRepo := s.repoFactory.NewUserRepo(s.db)

	// Find the user by email
	user, err := userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// Check if the password is correct
	passwordMatch, err := argon2ComparePlaintextAndHash(password, user.PasswordDigest)
	if err != nil {
		return nil, err
	}

	if !passwordMatch {
		return nil, ErrInvalidCredentials
	}

	var createTokensResult *CreateTokensResult

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		sessionRepo := s.repoFactory.NewSessionRepo(tx)
		refreshTokenRepo := s.repoFactory.NewRefreshTokenRepo(tx)
		accessTokenRepo := s.repoFactory.NewAccessTokenRepo(tx)

		// Create a session
		createTokensResult_tx, err := s.createSession(
			ctx,
			sessionRepo,
			refreshTokenRepo,
			accessTokenRepo,
			user,
			userAgent,
			ipAddress,
		)

		if err != nil {
			return err
		}

		createTokensResult = createTokensResult_tx

		return nil
	}); err != nil {
		return nil, err
	}

	return createTokensResult, nil
}
