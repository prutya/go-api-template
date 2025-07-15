package authentication_service

import (
	"context"
	"database/sql"
	"time"

	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/repo"
)

// This function assumes that the user has already been authenticated either
// through password or an email verification token
func (s *authenticationService) createSession(
	ctx context.Context,
	sessionRepo repo.SessionRepo,
	refreshTokenRepo repo.RefreshTokenRepo,
	accessTokenRepo repo.AccessTokenRepo,
	user *models.User,
	userAgent string,
	ipAddress string,
) (*CreateTokensResult, error) {
	// Generate a session ID
	sessionId, err := generateUUID()
	if err != nil {
		return nil, err
	}

	// Calculate the session expiration time
	sessionExpiresAt := time.Now().Add(s.config.AuthenticationRefreshTokenTTL)

	// Create a session in the database
	err = sessionRepo.Create(
		ctx,
		sessionId,
		user.ID,
		userAgent,
		ipAddress,
		sessionExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return s.createTokens(
		ctx,
		refreshTokenRepo,
		accessTokenRepo,
		user.ID,
		sessionId,
		sql.NullString{},
		sessionExpiresAt,
	)
}
