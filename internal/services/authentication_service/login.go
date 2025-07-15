package authentication_service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

// NOTE: I am not using transactions here, because it's fine if there's a
// session, refresh token or access token inserted without the rest of the
// records
func (s *authenticationService) Login(
	ctx context.Context,
	email string,
	password string,
	userAgent string,
	ipAddress string,
) (*CreateTokensResult, error) {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	// Normalize the email
	email = normalizeEmail(email)

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
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
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
