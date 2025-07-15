package authentication_service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"prutya/go-api-template/internal/repo"
)

type CreateTokensResult struct {
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	AccessToken           string
	CSRFToken             string
}

// This function assumes that the user has already been authenticated either
// through password or an email verification token
func (s *authenticationService) createTokens(
	ctx context.Context,
	refreshTokenRepo repo.RefreshTokenRepo,
	accessTokenRepo repo.AccessTokenRepo,
	userId string,
	sessionId string,
	parentRefreshTokenId sql.NullString,
	refreshTokenExpiresAt time.Time,
) (*CreateTokensResult, error) {
	// Generate a refresh token ID
	refreshTokenId, err := generateUUID()
	if err != nil {
		return nil, err
	}

	// Generate a refresh token secret
	refreshTokenSecret, err := generateSecret(s.config.AuthenticationRefreshTokenSecretLength)
	if err != nil {
		return nil, err
	}

	// Create a refresh token in the database
	if err := refreshTokenRepo.Create(
		ctx,
		refreshTokenId,
		sessionId,
		parentRefreshTokenId,
		refreshTokenSecret,
		refreshTokenExpiresAt,
	); err != nil {
		return nil, err
	}

	// Create an AccessToken
	accessTokenId, err := generateUUID()
	if err != nil {
		return nil, err
	}

	// Generate a new access token secret
	accessTokenSecret, err := generateSecret(s.config.AuthenticationAccessTokenSecretLength)
	if err != nil {
		return nil, err
	}

	// Set the expiration time for the access token
	accessTokenExpiresAt := time.Now().Add(s.config.AuthenticationAccessTokenTTL)

	// Create an access token in the database
	if err := accessTokenRepo.Create(
		ctx,
		accessTokenId,
		refreshTokenId,
		accessTokenSecret,
		accessTokenExpiresAt,
	); err != nil {
		return nil, err
	}

	// TODO: Extract the CSRF, JWT generation into separate functions?

	// Generate a CSRF token
	csrfToken, err := generateSecret(s.config.AuthenticationCSRFTokenLength)
	if err != nil {
		return nil, err
	}

	// Encode the CSRF token as a hex string so that it can be used in a JWT
	// NOTE: Why not Base64? I have encountered issues with Base64 encoding
	// between client and server in the past, so I prefer hex.
	csrfTokenString := hex.EncodeToString(csrfToken)

	// Create a JWT for the refresh token
	refreshTokenClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenId,
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    userId,
		CSRFToken: csrfTokenString,
	}
	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshTokenJWT.SignedString(refreshTokenSecret)
	if err != nil {
		return nil, err
	}

	// Create a JWT for the access token
	accessTokenClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenId,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userId,
	}
	accessTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessTokenJWT.SignedString(accessTokenSecret)
	if err != nil {
		return nil, err
	}

	return &CreateTokensResult{
		RefreshToken:          refreshTokenString,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		AccessToken:           accessTokenString,
		CSRFToken:             csrfTokenString,
	}, nil
}
