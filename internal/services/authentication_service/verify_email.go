package authentication_service

import (
	"context"
	"database/sql"
	"errors"
	"prutya/go-api-template/internal/argon2_utils"
	"prutya/go-api-template/internal/logger"
	"time"
)

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
			logger.DebugContext(ctx, ErrUserNotFound.Error(), "email", email)

			return nil, ErrUserNotFound
		}

		return nil, err
	}

	// Check number of attempts
	if user.EmailVerificationOtpAttempts >= s.config.AuthenticationEmailVerificationMaxAttempts {
		logger.DebugContext(ctx, ErrTooManyOTPAttempts.Error(), "user_id", user.ID)

		return nil, ErrTooManyOTPAttempts
	}

	// Check expiration
	if !user.EmailVerificationExpiresAt.Valid {
		logger.DebugContext(ctx, ErrEmailVerificationNotRequested.Error(), "user_id", user.ID)

		// This should not be null at this point
		return nil, ErrEmailVerificationNotRequested
	}

	if user.EmailVerificationExpiresAt.Time.Before(time.Now().UTC()) {
		logger.DebugContext(ctx, ErrEmailVerificationExpired.Error(), "user_id", user.ID)

		return nil, ErrEmailVerificationExpired
	}

	otpOk, err := argon2_utils.Compare(otp, user.EmailVerificationOtpDigest)
	if err != nil {
		return nil, err
	}

	if !otpOk {
		logger.DebugContext(ctx, ErrInvalidOTP.Error(), "user_id", user.ID, "otp", otp)

		if err := userRepo.IncrementEmailVerificationAttempts(ctx, user.ID); err != nil {
			return nil, err
		}

		return nil, ErrInvalidOTP
	}

	if user.EmailVerifiedAt.Valid {
		logger.DebugContext(ctx, ErrEmailAlreadyVerified.Error(), "user_id", user.ID)

		return nil, ErrEmailAlreadyVerified
	}

	if err := userRepo.CompleteEmailVerification(ctx, user.ID); err != nil {
		return nil, err
	}

	// Log the user in
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)
	refreshTokenRepo := s.repoFactory.NewRefreshTokenRepo(s.db)
	accessTokenRepo := s.repoFactory.NewAccessTokenRepo(s.db)

	return s.createSession(ctx, sessionRepo, refreshTokenRepo, accessTokenRepo, user, userAgent, ipAddress)
}
