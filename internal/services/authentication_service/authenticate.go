package authentication_service

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"

	"github.com/golang-jwt/jwt/v5"

	"prutya/go-api-template/internal/logger"
)

func (s *authenticationService) Authenticate(
	ctx context.Context,
	accessToken string,
) (*AccessTokenClaims, error) {
	logger := logger.MustFromContext(ctx)
	accessTokenRepo := s.repoFactory.NewAccessTokenRepo(s.db)

	// Parse the token
	parsedToken, err := jwt.ParseWithClaims(accessToken, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Extract the claims
		claims, ok := token.Claims.(*AccessTokenClaims)
		if !ok {
			return nil, ErrInvalidAccessTokenClaims
		}

		// Find the access token by ID
		//
		// NOTE: In a scenario when the Relying Party (RP) and the
		// Authorization Server (AS) are separate, this should be replaced with
		// validation of the token based on the public key of the AS.
		dbAccessToken, err := accessTokenRepo.FindById(ctx, claims.ID)
		if err != nil {
			logger.WarnContext(ctx, "AccessToken not found", "access_token_id", claims.ID)

			return nil, ErrAccessTokenNotFound
		}

		publicKey, err := x509.ParsePKIXPublicKey(dbAccessToken.PublicKey)
		if err != nil {
			return nil, err
		}

		return publicKey.(*ecdsa.PublicKey), nil
	}, jwt.WithValidMethods([]string{"ES256"}), jwt.WithExpirationRequired())

	if err != nil {
		if errors.Is(err, ErrInvalidAccessTokenClaims) || errors.Is(err, ErrAccessTokenNotFound) {
			return nil, err
		}

		return nil, ErrInvalidAccessToken
	}

	claims, ok := parsedToken.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, ErrInvalidAccessTokenClaims
	}

	return claims, nil
}
