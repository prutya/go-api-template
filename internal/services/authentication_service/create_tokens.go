package authentication_service

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"prutya/go-api-template/internal/repo"
)

type CreateTokensResult struct {
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	AccessToken           string
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
	refreshTokenSecret, err := generateRandomBytes(s.config.AuthenticationRefreshTokenSecretLength)
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
	accessTokenSecret, err := generateRandomBytes(s.config.AuthenticationAccessTokenSecretLength)
	if err != nil {
		return nil, err
	}

	// Set the expiration time for the access token
	accessTokenExpiresAt := time.Now().UTC().Add(s.config.AuthenticationAccessTokenTTL)

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

	// TODO: Extract JWT generation into separate functions?

	// Create a JWT for the refresh token
	refreshTokenClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenId,
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userId,
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
	}, nil
}
