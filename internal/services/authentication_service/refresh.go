package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"

	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
)

func (s *authenticationService) Refresh(ctx context.Context, refreshToken string, clientCSRFToken string) (*CreateTokensResult, error) {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	logger := loggerpkg.MustFromContext(ctx)

	refreshTokenRepo := s.repoFactory.NewRefreshTokenRepo(s.db)
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	var refreshTokenClaims *RefreshTokenClaims
	var dbRefreshToken *models.RefreshToken

	// Parse the token

	_, err := jwt.ParseWithClaims(refreshToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Extract the claims
		refreshTokenClaims_inner, ok := token.Claims.(*RefreshTokenClaims)
		if !ok {
			return nil, ErrInvalidRefreshTokenClaims
		}

		refreshTokenClaims = refreshTokenClaims_inner

		// Find the refresh token by ID
		dbRefreshToken_inner, err := refreshTokenRepo.FindById(ctx, refreshTokenClaims_inner.ID)
		if err != nil {
			logger.WarnContext(ctx, "RefreshToken not found", "refresh_token_id", refreshTokenClaims_inner.ID)

			return nil, ErrRefreshTokenNotFound
		}

		dbRefreshToken = dbRefreshToken_inner

		return dbRefreshToken_inner.Secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

	if err != nil {
		if errors.Is(err, ErrInvalidRefreshTokenClaims) || errors.Is(err, ErrRefreshTokenNotFound) {
			return nil, err
		}

		return nil, ErrRefreshTokenInvalid
	}

	// Check that CSRF tokens from the cookie and from the request match
	//
	// We are using the "double-submit" technique here. On login, we generate a
	// CSRF token and store it in the cookie. We also return it in the response
	// body. On the client side, we need to send the CSRF token in the request
	// body (or headers) and the cookie will be sent as well.
	//
	// When the refresh endpoint is called, we check that the CSRF token from the
	// cookie and the CSRF token from the request match. This works, because the
	// potential attacker can only trick the user into sending cookies, but only
	// the web client (JS) can set the CSRF token in the request body.

	if refreshTokenClaims.CSRFToken != clientCSRFToken {
		logger.WarnContext(ctx, "CSRF token mismatch", "refresh_token_id", refreshTokenClaims.ID)

		return nil, ErrCSRFTokenMismatch
	}

	// Check if the refresh token is revoked

	if dbRefreshToken.RevokedAt.Valid {
		// Check the leeway period (grace period) for the refresh token
		// This allows to use the refresh token for a short period after it has been
		// revoked to prevent race conditions when multiple refresh requests are
		// sent at the same time

		if time.Now().After(dbRefreshToken.LeewayExpiresAt.Time) {
			logger.WarnContext(ctx, "RefreshToken reuse detected", "refresh_token_id", dbRefreshToken.ID)

			// The session is compromised, so we need to terminate it
			if err := sessionRepo.TerminateByID(ctx, dbRefreshToken.SessionID, time.Now()); err != nil {
				logger.ErrorContext(ctx, "Failed to terminate session", "session_id", dbRefreshToken.SessionID, "error", err)

				return nil, err
			}

			return nil, ErrRefreshTokenRevoked
		} else {
			logger.InfoContext(
				ctx,
				"RefreshToken reuse detected but within the leeway period",
				"refresh_token_id", dbRefreshToken.ID,
			)
		}
	}

	// Check if the session is terminated

	// NOTE: I am not checking if the session has EXPIRED, because this check is
	// already done on the refresh token, which has the same expires_at

	session, err := sessionRepo.FindByID(ctx, dbRefreshToken.SessionID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Session not found", "session_id", dbRefreshToken.SessionID)

			return nil, ErrSessionNotFound
		}

		return nil, err
	}

	if session.TerminatedAt.Valid {
		logger.WarnContext(ctx, "Session already terminated", "session_id", dbRefreshToken.SessionID)

		return nil, ErrSessionAlreadyTerminated
	}

	// NOTE: I am not using transactions here, because the revoked token still has
	// a leeway time when it can be used again

	// Revoke the old refresh token

	revokedAt := time.Now()
	leewayExpiresAt := revokedAt.Add(s.config.AuthenticationRefreshTokenLeeway)

	if err := refreshTokenRepo.Revoke(ctx, dbRefreshToken.ID, revokedAt, leewayExpiresAt); err != nil {
		return nil, err
	}

	var createTokensResult *CreateTokensResult

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		sessionRepoTx := s.repoFactory.NewSessionRepo(tx)
		refreshTokenRepoTx := s.repoFactory.NewRefreshTokenRepo(tx)
		accessTokenRepoTx := s.repoFactory.NewAccessTokenRepo(tx)

		newSessionExpiresAt := time.Now().Add(s.config.AuthenticationRefreshTokenTTL)

		// Create new tokens
		createTokensResult_tx, err := s.createTokens(
			ctx,
			refreshTokenRepoTx,
			accessTokenRepoTx,
			session.UserID,
			session.ID,
			sql.NullString{String: dbRefreshToken.ID, Valid: true},
			newSessionExpiresAt,
		)
		if err != nil {
			return err
		}

		// Update session's expires_at
		if err := sessionRepoTx.UpdateExpiresAtByID(ctx, dbRefreshToken.SessionID, newSessionExpiresAt); err != nil {
			return err
		}

		createTokensResult = createTokensResult_tx

		return nil
	}); err != nil {
		return nil, err
	}

	return createTokensResult, nil
}
