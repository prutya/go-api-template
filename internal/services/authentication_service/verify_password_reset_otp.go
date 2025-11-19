package authentication_service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"prutya/go-api-template/internal/logger"
)

func (s *authenticationService) VerifyPasswordResetOTP(
	ctx context.Context,
	email string,
	otp string,
) (string, error) {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	logger := logger.MustFromContext(ctx)
	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.DebugContext(ctx, ErrUserNotFound.Error(), "email", email)

			return "", ErrUserNotFound
		}

		return "", err
	}

	// Check number of attempts
	if user.PasswordResetOtpAttempts >= s.config.AuthenticationPasswordResetMaxAttempts {
		logger.DebugContext(ctx, ErrTooManyOTPAttempts.Error(), "user_id", user.ID)

		return "", ErrTooManyOTPAttempts
	}

	// Check expiration
	if !user.PasswordResetExpiresAt.Valid {
		logger.DebugContext(ctx, ErrPasswordResetNotRequested.Error(), "user_id", user.ID)

		return "", ErrPasswordResetNotRequested
	}

	if user.PasswordResetExpiresAt.Time.Before(time.Now().UTC()) {
		logger.DebugContext(ctx, ErrPasswordResetExpired.Error(), "user_id", user.ID)

		return "", ErrPasswordResetExpired
	}

	// Verify OTP
	hmacOk, err := checkHmac([]byte(otp), s.config.AuthenticationOtpHmacSecret, user.PasswordResetOtpHmac)
	if err != nil {
		return "", err
	}

	if !hmacOk {
		logger.DebugContext(ctx, "Invalid OTP", "user_id", user.ID, "otp", otp)

		if err := userRepo.IncrementPasswordResetAttempts(ctx, user.ID); err != nil {
			return "", err
		}

		return "", ErrInvalidOTP
	}

	// Generate reset token key pair
	resetTokenPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	resetTokenPublicKeyBytes, err := x509.MarshalPKIXPublicKey(&resetTokenPrivateKey.PublicKey)
	if err != nil {
		return "", err
	}

	// Store public key and clean otp state
	if err := userRepo.StorePasswordResetTokenKey(ctx, user.ID, resetTokenPublicKeyBytes); err != nil {
		return "", err
	}

	// Build a JWT
	tokenExpiresAt := time.Now().UTC().Add(s.config.AuthenticationPasswordResetTokenTTL)
	tokenClaims := PasswordResetTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		UserID: user.ID,
	}
	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodES256, tokenClaims)
	tokenString, err := tokenJWT.SignedString(resetTokenPrivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
