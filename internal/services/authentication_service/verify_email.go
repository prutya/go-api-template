package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"prutya/go-api-template/internal/logger"
	"time"
)

var ErrInvalidEmailVerificationTokenClaims = errors.New("invalid email verification token claims")
var ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
var ErrEmailVerificationTokenInvalid = errors.New("email verification token invalid")

func (s *authenticationService) VerifyEmail(
	ctx context.Context,
	email string,
	otp string,
	userAgent string,
	ipAddress string,
) (*CreateTokensResult, error) {
	defer withMinimumAllowedFunctionDuration(ctx, s.config.AuthenticationTimingAttackDelay)()

	logger := logger.MustFromContext(ctx)
	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.DebugContext(ctx, "user not found", "email", email)

			return nil, ErrUserNotFound
		}

		return nil, err
	}

	// Check number of attempts
	if user.EmailVerificationOtpAttempts >= s.config.AuthenticationEmailVerificationMaxAttempts {
		logger.DebugContext(ctx, "Too many email verification attempts", "user_id", user.ID)

		return nil, ErrTooManyOTPAttempts
	}

	// Check expiration
	if !user.EmailVerificationExpiresAt.Valid {
		logger.WarnContext(ctx, "User's email_verification_expires_at was null", "user_id", user.ID)

		// This should not be null at this point
		return nil, ErrEmailVerificationExpired
	}

	if user.EmailVerificationExpiresAt.Time.Before(time.Now().UTC()) {
		logger.DebugContext(ctx, "Email verification expired", "user_id", user.ID)

		return nil, ErrEmailVerificationExpired
	}

	// Verify OTP
	// TODO: Check if hmac is null in db???
	hmacOk, err := checkHmac([]byte(otp), s.config.AuthenticationOtpHmacSecret, user.EmailVerificationOtpHmac)
	if err != nil {
		return nil, err
	}

	if !hmacOk {
		logger.DebugContext(ctx, "Invalid OTP", "user_id", user.ID, "otp", otp)

		// TODO: This looks like it can lead to a lost update if another transaction
		// is incrementing the attempts at the same time. We can potentially end up
		// with one or more extra attempts
		// Increment attempts
		if err := userRepo.IncrementEmailVerificationAttempts(ctx, user.ID); err != nil {
			return nil, err
		}

		return nil, ErrInvalidOTP
	}

	if user.EmailVerifiedAt.Valid {
		return nil, ErrEmailAlreadyVerified
	}

	if err := userRepo.CompleteEmailVerification(ctx, user.ID); err != nil {
		return nil, err
	}

	// Create a new session for the user
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)
	refreshTokenRepo := s.repoFactory.NewRefreshTokenRepo(s.db)
	accessTokenRepo := s.repoFactory.NewAccessTokenRepo(s.db)

	return s.createSession(ctx, sessionRepo, refreshTokenRepo, accessTokenRepo, user, userAgent, ipAddress)
}
