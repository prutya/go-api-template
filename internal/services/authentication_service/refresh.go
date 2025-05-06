package authentication_service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"

	"prutya/go-api-template/internal/logger"
	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
)

func (s *authenticationService) Refresh(ctx context.Context, refreshToken string) (string, time.Time, string, error) {
	// Prevent the potential attacker from measuring the response time
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		timeLeft := s.config.AuthenticationTimingAttackDelay - duration

		if timeLeft > 0 {
			time.Sleep(timeLeft)
		}
	}()

	logger := logger.MustFromContext(ctx)

	var dbRefreshToken *models.RefreshToken

	// Parse the token

	_, err := jwt.ParseWithClaims(refreshToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Extract the claims
		refreshTokenClaims_inner, ok := token.Claims.(*RefreshTokenClaims)
		if !ok {
			return nil, ErrInvalidRefreshTokenClaims
		}

		// Find the refresh token by ID
		//
		// NOTE: In a scenario when the Relying Party (RP) and the
		// Authorization Server (AS) are separate, this should be replaced with a
		// validation of the token based on the public key of the AS.
		// Find the refresh token by ID
		dbRefreshToken_inner, err := s.refreshTokenRepo.FindById(ctx, refreshTokenClaims_inner.ID)
		if err != nil {
			logger.WarnContext(ctx, "RefreshToken not found", "refresh_token_id", refreshTokenClaims_inner.ID)

			return nil, ErrRefreshTokenNotFound
		}

		dbRefreshToken = dbRefreshToken_inner

		return dbRefreshToken_inner.Secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

	if err != nil {
		if errors.Is(err, ErrInvalidRefreshTokenClaims) || errors.Is(err, ErrRefreshTokenNotFound) {
			return "", time.Time{}, "", err
		}

		return "", time.Time{}, "", ErrRefreshTokenInvalid
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
			if err := s.sessionRepo.TerminateByID(ctx, dbRefreshToken.SessionID, time.Now()); err != nil {
				logger.ErrorContext(ctx, "Failed to terminate session", "session_id", dbRefreshToken.SessionID, "error", err)

				return "", time.Time{}, "", err
			}

			return "", time.Time{}, "", ErrRefreshTokenRevoked
		} else {
			logger.InfoContext(
				ctx,
				"RefreshToken reuse detected but within the leeway period",
				"refresh_token_id", dbRefreshToken.ID,
			)
		}
	}

	// Check if the session is terminated

	session, err := s.sessionRepo.FindById(ctx, dbRefreshToken.SessionID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Session not found", "session_id", dbRefreshToken.SessionID)

			return "", time.Time{}, "", ErrSessionNotFound
		}

		return "", time.Time{}, "", err
	}

	if session.TerminatedAt.Valid {
		logger.WarnContext(ctx, "Session terminated", "session_id", dbRefreshToken.SessionID)

		return "", time.Time{}, "", ErrSessionTerminated
	}

	// Create a new RefreshToken

	newRefreshTokenUUID, err := uuid.NewV7()

	if err != nil {
		return "", time.Time{}, "", err
	}

	newRefreshTokenId := newRefreshTokenUUID.String()

	newRefreshTokenSecret := make([]byte, s.config.AuthenticationRefreshTokenSecretLength)

	_, err = rand.Read(newRefreshTokenSecret)

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Make sure that the refresh token secret is redacted in logs
	ctx = loggerpkg.NewContextWithRedactedSecret(ctx, hex.EncodeToString(newRefreshTokenSecret))

	newRefreshTokenExpiresAt := time.Now().Add(s.config.AuthenticationRefreshTokenTTL)

	if err := s.refreshTokenRepo.Create(
		ctx,
		newRefreshTokenId,
		session.ID,
		sql.NullString{Valid: true, String: dbRefreshToken.ID},
		newRefreshTokenSecret,
		newRefreshTokenExpiresAt,
	); err != nil {
		return "", time.Time{}, "", err
	}

	// Create a new AccessToken

	newAccessTokenUUID, err := uuid.NewV7()

	if err != nil {
		return "", time.Time{}, "", err
	}

	newAccessTokenId := newAccessTokenUUID.String()

	newAccessTokenSecret := make([]byte, s.config.AuthenticationAccessTokenSecretLength)

	_, err = rand.Read(newAccessTokenSecret)

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Make sure that the access token secret is redacted in logs
	ctx = loggerpkg.NewContextWithRedactedSecret(ctx, hex.EncodeToString(newAccessTokenSecret))

	newAccessTokenExpiresAt := time.Now().Add(s.config.AuthenticationAccessTokenTTL)

	if err := s.accessTokenRepo.Create(
		ctx,
		newAccessTokenId,
		newRefreshTokenId,
		newAccessTokenSecret,
		newAccessTokenExpiresAt,
	); err != nil {
		return "", time.Time{}, "", err
	}

	// Create a JWT for the new refresh token

	newRefreshTokenClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newRefreshTokenId,
			ExpiresAt: jwt.NewNumericDate(newRefreshTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: session.UserID,
	}

	newRefreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, newRefreshTokenClaims)

	newRefreshTokenString, err := newRefreshTokenJWT.SignedString(newRefreshTokenSecret)
	if err != nil {
		return "", time.Time{}, "", err
	}

	// Create a JWT for the new access token

	newAccessTokenClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newAccessTokenId,
			ExpiresAt: jwt.NewNumericDate(newAccessTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: session.UserID,
	}

	newAccessTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessTokenClaims)

	newAccessTokenString, err := newAccessTokenJWT.SignedString(newAccessTokenSecret)

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Revoke the old refresh token

	revokedAt := time.Now()
	leewayExpiresAt := revokedAt.Add(s.config.AuthenticationRefreshTokenLeeway)

	if err := s.refreshTokenRepo.Revoke(ctx, dbRefreshToken.ID, revokedAt, leewayExpiresAt); err != nil {
		return "", time.Time{}, "", err
	}

	return newRefreshTokenString, newRefreshTokenExpiresAt, newAccessTokenString, nil
}
