package authentication_service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
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

	// Generate refresh token key pair
	refreshTokenPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	refreshTokenPublicKeyBytes, err := x509.MarshalPKIXPublicKey(&refreshTokenPrivateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	// Create a refresh token in the database
	if err := refreshTokenRepo.Create(
		ctx,
		refreshTokenId,
		sessionId,
		parentRefreshTokenId,
		refreshTokenPublicKeyBytes,
		refreshTokenExpiresAt,
	); err != nil {
		return nil, err
	}

	// Create an AccessToken
	accessTokenId, err := generateUUID()
	if err != nil {
		return nil, err
	}

	// Generate access token key pair
	accessTokenPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	accessTokenPublicKeyBytes, err := x509.MarshalPKIXPublicKey(&accessTokenPrivateKey.PublicKey)
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
		accessTokenPublicKeyBytes,
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
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		UserID: userId,
	}
	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodES256, refreshTokenClaims)
	refreshTokenString, err := refreshTokenJWT.SignedString(refreshTokenPrivateKey)
	if err != nil {
		return nil, err
	}

	// Create a JWT for the access token
	accessTokenClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenId,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		UserID: userId,
	}
	accessTokenJWT := jwt.NewWithClaims(jwt.SigningMethodES256, accessTokenClaims)
	accessTokenString, err := accessTokenJWT.SignedString(accessTokenPrivateKey)
	if err != nil {
		return nil, err
	}

	return &CreateTokensResult{
		RefreshToken:          refreshTokenString,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		AccessToken:           accessTokenString,
	}, nil
}
