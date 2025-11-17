package authentication_service

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	text_template "text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var PasswordResetEmailTemplateText = text_template.Must(text_template.New("password_reset_email").Parse(
	`
	Hi!
	To reset your password, please use the link below:
	{{.ResetURL}}
	This link will expire in {{.TokenTTL}}.
	If you did not request a password reset, please ignore this email.
	`,
))

var PasswordResetEmailTemplateHTML = text_template.Must(text_template.New("password_reset_email").Parse(
	`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Reset your password</title>
	</head>
	<body>
		<p>Hi!</p>
		<p>To reset your password, please use the link below:</p>
		<p><a href="{{.ResetURL}}">Reset your password</a></p>
		<p>This link will expire in {{.TokenTTL}}.</p>
		<p>If you did not request a password reset, please ignore this email.</p>
	</body>
	</html>
	`,
))

var ErrPasswordResetRateLimited = errors.New("password reset rate limited")

func (s *authenticationService) SendPasswordResetEmail(ctx context.Context, userID string) error {
	passwordResetTokenRepo := s.repoFactory.NewPasswordResetTokenRepo(s.db)
	userRepo := s.repoFactory.NewUserRepo(s.db)

	// Fetch the user
	user, err := findUserByID(ctx, userRepo, userID)
	if err != nil {
		return err
	}

	// // Check if password reset is rate limited
	// if user.PasswordResetRateLimitedUntil.Valid {
	// 	if user.PasswordResetRateLimitedUntil.Time.After(time.Now()) {
	// 		logger.MustFromContext(ctx).WarnContext(ctx, "password reset rate limited", "user_id", user.ID, "error", err)

	// 		return ErrPasswordResetRateLimited
	// 	}
	// }

	// Generate a new password reset token and store it
	tokenID, err := generateUUID()
	if err != nil {
		return err
	}

	tokenSecret, err := generateRandomBytes(s.config.AuthenticationPasswordResetTokenSecretLength)
	if err != nil {
		return err
	}

	tokenExpiresAt := time.Now().UTC().Add(s.config.AuthenticationPasswordResetTokenTTL)

	if err := passwordResetTokenRepo.Create(ctx, tokenID, userID, tokenSecret, tokenExpiresAt); err != nil {
		return err
	}

	// Build a token JWT

	tokenClaims := &PasswordResetTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(tokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	tokenString, err := tokenJWT.SignedString(tokenSecret)
	if err != nil {
		return err
	}

	// Prepare the reset URL
	resetURL := s.config.AuthenticationPasswordResetURL
	resetURL = resetURL + "?" + url.Values{"token": []string{tokenString}}.Encode()

	// Render the email template

	var textContentBuf bytes.Buffer
	if err := PasswordResetEmailTemplateText.Execute(&textContentBuf, map[string]string{
		"ResetURL": resetURL,
		"TokenTTL": s.config.AuthenticationPasswordResetTokenTTL.String(),
	}); err != nil {
		return err
	}

	var htmlContentBuf bytes.Buffer
	if err := PasswordResetEmailTemplateHTML.Execute(&htmlContentBuf, map[string]string{
		"ResetURL": resetURL,
		"TokenTTL": s.config.AuthenticationPasswordResetTokenTTL.String(),
	}); err != nil {
		return err
	}

	// Send the email
	if err := s.transactionalEmailService.SendEmail(
		ctx,
		user.Email,
		user.ID,
		"Password reset",
		textContentBuf.String(),
		htmlContentBuf.String(),
	); err != nil {
		return err
	}

	// Update the sent_at of the password reset token
	if err := passwordResetTokenRepo.UpdateSentAt(ctx, tokenID, time.Now()); err != nil {
		return err
	}

	// Update the password reset rate limit
	// NOTE: This is not in the same transaction as UpdateSentAt, because it does
	// not have to be 100% correct. Most of the time it should work.
	if err := userRepo.UpdatePasswordResetRateLimit(
		ctx,
		user.ID,
		time.Now().UTC().Add(s.config.AuthenticationPasswordResetRateLimitInterval),
	); err != nil {
		return err
	}

	return nil
}
