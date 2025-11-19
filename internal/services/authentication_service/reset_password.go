package authentication_service

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"errors"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: Move errors from the common file closer to each method to reduce
// WTFs per second when reading this code

func (s *authenticationService) ResetPassword(
	ctx context.Context,
	token string,
	newPassword string,
	userAgent string,
	ipAddress string,
) (*CreateTokensResult, error) {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	logger := logger.MustFromContext(ctx)
	userRepo := s.repoFactory.NewUserRepo(s.db)
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	var user *models.User

	// Prepare the validation key function
	keyFunc := func(token *jwt.Token) (any, error) {
		// Extract the claims
		claims, ok := token.Claims.(*PasswordResetTokenClaims)
		if !ok {
			return nil, ErrInvalidPasswordResetTokenClaims
		}

		// Find the user by ID
		user_keyfunc, err := userRepo.FindByID(ctx, claims.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.DebugContext(ctx, ErrUserNotFound.Error(), "user_id", claims.UserID)

				return "", ErrUserNotFound
			}

			return nil, err
		}
		user = user_keyfunc

		publicKey, err := x509.ParsePKIXPublicKey(user.PasswordResetTokenPublicKey)
		if err != nil {
			return nil, err
		}

		return publicKey.(*ecdsa.PublicKey), nil
	}

	// Validate the token
	if _, err := jwt.ParseWithClaims(
		token,
		&PasswordResetTokenClaims{},
		keyFunc,
		jwt.WithValidMethods([]string{"ES256"}),
		jwt.WithExpirationRequired(),
	); err != nil {
		logger.WarnContext(ctx, "Password reset token verification failed", "error", err.Error())
		logger.DebugContext(ctx, "Password reset token verification failed", "password_reset_token", token)

		return nil, ErrInvalidPasswordResetToken
	}

	// Terminate all sessions for the user
	if err := sessionRepo.TerminateAllSessions(ctx, user.ID); err != nil {
		return nil, err
	}

	// Hash the new password
	newPasswordDigest, err := s.argon2GenerateHashFromPassword(newPassword)
	if err != nil {
		return nil, err
	}

	// Update the password and invalidate the token
	if err := userRepo.ResetPassword(ctx, user.ID, newPasswordDigest); err != nil {
		return nil, err
	}

	// Log the user in
	refreshTokenRepo := s.repoFactory.NewRefreshTokenRepo(s.db)
	accessTokenRepo := s.repoFactory.NewAccessTokenRepo(s.db)

	return s.createSession(ctx, sessionRepo, refreshTokenRepo, accessTokenRepo, user, userAgent, ipAddress)
}
