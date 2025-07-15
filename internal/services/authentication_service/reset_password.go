package authentication_service

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
)

var ErrInvalidResetPasswordTokenClaims = errors.New("invalid reset password token claims")
var ErrResetPasswordTokenNotFound = errors.New("reset password token not found")
var ErrResetPasswordTokenInvalid = errors.New("reset password token invalid")

func (s *authenticationService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	defer withMinimumAllowedFunctionDuration(s.config.AuthenticationTimingAttackDelay)()

	passwordResetTokenRepo := s.repoFactory.NewPasswordResetTokenRepo(s.db)
	userRepo := s.repoFactory.NewUserRepo(s.db)
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	// Parse the token

	var dbPasswordResetToken *models.PasswordResetToken

	_, err := jwt.ParseWithClaims(token, &PasswordResetTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Extract the claims
		passwordResetTokenClaims_inner, ok := token.Claims.(*PasswordResetTokenClaims)
		if !ok {
			return nil, ErrInvalidResetPasswordTokenClaims
		}

		// Find the password reset token by ID
		dbPasswordResetToken_inner, err := passwordResetTokenRepo.FindByID(ctx, passwordResetTokenClaims_inner.ID)
		if err != nil {
			return nil, ErrResetPasswordTokenNotFound
		}

		dbPasswordResetToken = dbPasswordResetToken_inner

		return dbPasswordResetToken_inner.Secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

	if err != nil {
		if errors.Is(err, ErrInvalidResetPasswordTokenClaims) || errors.Is(err, ErrResetPasswordTokenNotFound) {
			return err
		}

		return ErrResetPasswordTokenInvalid
	}

	// Check if the token has already been used
	if dbPasswordResetToken.ResetAt.Valid {
		logger := logger.MustFromContext(ctx)

		logger.WarnContext(
			ctx,
			"password reset token already used",
			"password_reset_token_id", dbPasswordResetToken.ID,
		)

		return ErrResetPasswordTokenInvalid
	}

	// Find the user by ID
	dbUser, err := userRepo.FindByID(ctx, dbPasswordResetToken.UserID)
	if err != nil {
		return err
	}

	// Hash the new password
	newPasswordDigest, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.AuthenticationBcryptCost)
	if err != nil {
		return err
	}

	// NOTE: The token can only be used once, so we should make sure that it's
	// applied properly via a transaction

	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := s.repoFactory.NewUserRepo(tx)
		passwordResetTokenRepo := s.repoFactory.NewPasswordResetTokenRepo(tx)

		// Update the user's password
		if err := userRepo.UpdatePasswordDigest(ctx, dbUser.ID, string(newPasswordDigest)); err != nil {
			return err
		}

		// Mark the password reset token as used
		if err := passwordResetTokenRepo.MarkAsReset(ctx, dbPasswordResetToken.ID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Terminate all sessions for the user
	// NOTE: This is not in the transaction, because it can be handled separately
	// manually in case it fails
	if err := sessionRepo.TerminateAllSessions(ctx, dbUser.ID); err != nil {
		return err
	}

	return nil
}
