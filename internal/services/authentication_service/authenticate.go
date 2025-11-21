package authentication_service

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"errors"
	"prutya/go-api-template/internal/logger"

	"github.com/golang-jwt/jwt/v5"
)

func (s *authenticationService) Authenticate(
	ctx context.Context,
	accessToken string,
) (*AccessTokenClaims, error) {
	logger := logger.MustFromContext(ctx)
	accessTokenRepo := s.repoFactory.NewAccessTokenRepo(s.db)

	// Prepare the validation key function
	keyFunc := func(token *jwt.Token) (any, error) {
		// Extract the claims
		claims, ok := token.Claims.(*AccessTokenClaims)
		if !ok {
			return nil, ErrInvalidAccessTokenClaims
		}

		// Find the access token by ID
		//
		// NOTE: In a scenario when the Relying Party (RP a.k.a. Resource Server,
		// in other words - one of your services that does not manage user's
		// sessions) and the Authorization Server (AS) are separate, this should be
		// replaced with validation of the token based on the public key from the
		// AS.
		//
		// TODO: This does not look scaleable because we need the public key from AS
		// and this will result in a network request. Since there is a new public
		// key for every JWT, there will be too many public keys to cache.
		// One way to mitigate it is to use a rotated public key on the AS, use it
		// to sign JWTs and fetch it once and cache it on RP. The downside of this
		// approach is that we lose granularity - i.e. we can't revoke just one
		// JWT – all JWTs signed by a public key will be revoked when the respective
		// public key is deleted.
		dbAccessToken, err := accessTokenRepo.FindById(ctx, claims.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.DebugContext(ctx, ErrAccessTokenNotFound.Error(), "access_token_id", claims.ID)

				return nil, ErrAccessTokenNotFound
			}

			return nil, err
		}

		publicKey, err := x509.ParsePKIXPublicKey(dbAccessToken.PublicKey)
		if err != nil {
			return nil, err
		}

		return publicKey.(*ecdsa.PublicKey), nil
	}

	claims := &AccessTokenClaims{}

	// Parse the token
	_, err := jwt.ParseWithClaims(
		accessToken,
		claims,
		keyFunc,
		jwt.WithValidMethods([]string{"ES256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		logger.WarnContext(ctx, "Access token verification failed", "error", err.Error())
		logger.DebugContext(ctx, "Access token verification failed", "access_token", accessToken)

		return nil, ErrInvalidAccessToken
	}

	return claims, nil
}
