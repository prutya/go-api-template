package authentication_service

import (
	"context"
	"errors"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
)

var ErrInvalidEmailVerificationTokenClaims = errors.New("invalid email verification token claims")
var ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
var ErrEmailVerificationTokenInvalid = errors.New("email verification token invalid")

func (s *authenticationService) VerifyEmail(
	ctx context.Context,
	token string,
	userAgent string,
	ipAddress string,
) (*CreateTokensResult, error) {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	emailVerificationTokenRepo := s.repoFactory.NewEmailVerificationTokenRepo(s.db)
	userRepo := s.repoFactory.NewUserRepo(s.db)
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)
	refreshTokenRepo := s.repoFactory.NewRefreshTokenRepo(s.db)
	accessTokenRepo := s.repoFactory.NewAccessTokenRepo(s.db)

	// Parse the email verification token
	var dbEmailVerificationToken *models.EmailVerificationToken
	_, err := jwt.ParseWithClaims(token, &EmailVerificationTokenClaims{}, func(token *jwt.Token) (any, error) {
		// Extract the claims
		emailVerificationTokenClaims_inner, ok := token.Claims.(*EmailVerificationTokenClaims)
		if !ok {
			return nil, ErrInvalidEmailVerificationTokenClaims
		}

		// Find the email verification token by ID
		dbEmailVerificationToken_inner, err := emailVerificationTokenRepo.FindByID(
			ctx, emailVerificationTokenClaims_inner.ID,
		)
		if err != nil {
			return nil, ErrEmailVerificationTokenNotFound
		}

		dbEmailVerificationToken = dbEmailVerificationToken_inner

		return dbEmailVerificationToken_inner.Secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, ErrInvalidEmailVerificationTokenClaims) || errors.Is(err, ErrEmailVerificationTokenNotFound) {
			return nil, err
		}

		return nil, ErrEmailVerificationTokenInvalid
	}

	// Check if the verification token has already been used
	if dbEmailVerificationToken.VerifiedAt.Valid {
		logger := logger.MustFromContext(ctx)

		logger.WarnContext(
			ctx,
			"verification token already used",
			"email_verification_token_id", dbEmailVerificationToken.ID,
		)

		return nil, ErrEmailVerificationTokenInvalid
	}

	// Find the user by ID
	user, err := findUserByID(ctx, userRepo, dbEmailVerificationToken.UserID)
	if err != nil {
		return nil, err
	}

	// Check if the email address is already verified
	if user.EmailVerifiedAt.Valid {
		return nil, ErrEmailAlreadyVerified
	}

	// NOTE: The token can only be used once, so we should make sure that it's
	// applied properly via a transaction

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := s.repoFactory.NewUserRepo(tx)
		emailVerificationTokenRepo := s.repoFactory.NewEmailVerificationTokenRepo(tx)

		// Mark the user's email address as verified
		err = userRepo.MarkEmailAsVerified(ctx, user.ID)
		if err != nil {
			return err
		}

		// Mark the verification code as verified
		err = emailVerificationTokenRepo.MarkAsVerified(ctx, dbEmailVerificationToken.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Create a new session for the user
	return s.createSession(ctx, sessionRepo, refreshTokenRepo, accessTokenRepo, user, userAgent, ipAddress)
}
