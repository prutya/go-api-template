package utils

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/authentication_service"
)

type accessTokenClaimsContextKeyType struct{}

var accessTokenClaimsContextKey = accessTokenClaimsContextKeyType{}

func NewAuthenticationMiddleware(authenticationService authentication_service.AuthenticationService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logger.MustFromContext(ctx)

			// Read the token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RenderError(w, r, ErrUnauthorized)
				return
			}

			tokenString := authHeader[len("Bearer "):]
			if tokenString == "" {
				RenderError(w, r, ErrUnauthorized)
				return
			}

			accessTokenClaims, err := authenticationService.Authenticate(ctx, tokenString)

			if err != nil {
				if errors.Is(err, authentication_service.ErrInvalidAccessTokenClaims) ||
					errors.Is(err, authentication_service.ErrAccessTokenNotFound) ||
					errors.Is(err, authentication_service.ErrInvalidAccessToken) {

					RenderError(w, r, ErrUnauthorized)
					return
				}

				RenderError(w, r, err)
				return
			}

			// Store the access token claims in the context
			ctx = NewContextWithAccessTokenClaims(ctx, accessTokenClaims)
			r = r.WithContext(ctx)

			logger.Info("User authenticated", zap.String("user_id", accessTokenClaims.UserID))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func NewContextWithAccessTokenClaims(ctx context.Context, claims *authentication_service.AccessTokenClaims) context.Context {
	return context.WithValue(ctx, accessTokenClaimsContextKey, claims)
}

func GetAccessTokenClaimsFromContext(ctx context.Context) *authentication_service.AccessTokenClaims {
	if claims, ok := ctx.Value(accessTokenClaimsContextKey).(*authentication_service.AccessTokenClaims); ok {
		return claims
	}

	panic("no access token claims in context")
}
