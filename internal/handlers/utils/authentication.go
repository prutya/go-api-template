// TODO: Test

package utils

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/services/authentication_service"
)

type currentUserContextKeyType struct{}

type currentSessionContextKeyType struct{}

var currentUserContextKey = currentUserContextKeyType{}

var currentSessionContextKey = currentSessionContextKeyType{}

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

			user, userSession, err := authenticationService.Authenticate(ctx, tokenString)

			if err != nil {
				if errors.Is(err, authentication_service.ErrInvalidTokenClaims) ||
					errors.Is(err, authentication_service.ErrSessionNotFound) ||
					errors.Is(err, authentication_service.ErrInvalidToken) ||
					errors.Is(err, authentication_service.ErrSessionExpired) ||
					errors.Is(err, authentication_service.ErrSessionTerminated) ||
					errors.Is(err, authentication_service.ErrUserNotFound) {

					RenderError(w, r, ErrUnauthorized)
					return
				}

				RenderError(w, r, err)
				return
			}

			// Store the current user and the session in the context
			ctx = NewContextWithCurrentUser(r.Context(), user)
			ctx = NewContextWithCurrentSession(ctx, userSession)
			r = r.WithContext(ctx)

			logger.Info("User authenticated", zap.String("user_id", user.ID), zap.String("session_id", userSession.ID))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func NewContextWithCurrentUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, currentUserContextKey, user)
}

func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(currentUserContextKey).(*models.User); ok {
		return user
	}

	panic("No user in context")
}

func NewContextWithCurrentSession(ctx context.Context, session *models.Session) context.Context {
	return context.WithValue(ctx, currentSessionContextKey, session)
}

func GetSessionFromContext(ctx context.Context) *models.Session {
	if session, ok := ctx.Value(currentSessionContextKey).(*models.Session); ok {
		return session
	}

	panic("No session in context")
}
