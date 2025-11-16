package authentication_service

import (
	"bytes"
	"context"
	"errors"
	html_template "html/template"
	"net/url"
	text_template "text/template"
	"time"

	"prutya/go-api-template/internal/logger"

	"github.com/golang-jwt/jwt/v5"
)

var VerificationEmailTemplateText = text_template.Must(text_template.New("verification_email").Parse(
	`
	Hi!
	Thank you for signing up! To complete your registration, please verify your email address via the link below:
	{{.VerificationURL}}
	This link will expire in {{.TokenTTL}}.
	If you did not sign up for this account, please ignore this email.
	`,
))

var VerificationEmailTemplateHTML = html_template.Must(html_template.New("verification_email").Parse(
	`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Verify your email address</title>
	</head>
	<body>
		<p>Hi!</p>
		<p>Thank you for signing up! To complete your registration, please verify your email address via the link below:</p>
		<p><a href="{{.VerificationURL}}">Verify your email address</a></p>
		<p>This link will expire in {{.TokenTTL}}.</p>
		<p>If you did not sign up for this account, please ignore this email.</p>
	</body>
	</html>
	`,
))

var ErrEmailVerificationRateLimited = errors.New("email verification rate limited")

func (s *authenticationService) SendVerificationEmail(ctx context.Context, userID string) error {
	logger := logger.MustFromContext(ctx)

	emailVerificationTokenRepo := s.repoFactory.NewEmailVerificationTokenRepo(s.db)
	userRepo := s.repoFactory.NewUserRepo(s.db)

	// Fetch the user
	user, err := findUserByID(ctx, userRepo, userID)
	if err != nil {
		return err
	}

	// Check if the user is already verified
	if user.EmailVerifiedAt.Valid {
		logger.InfoContext(ctx, "user already verified", "user_id", userID)
		return nil
	}

	// Check if email verification is rate limited
	if user.EmailVerificationRateLimitedUntil.Valid {
		if user.EmailVerificationRateLimitedUntil.Time.After(time.Now()) {
			logger.WarnContext(
				ctx,
				"email verification rate limited",
				"user_id", userID,
				"rate_limited_until", user.EmailVerificationRateLimitedUntil.Time,
			)

			return ErrEmailVerificationRateLimited
		}
	}

	// Generate a new verification token and store it

	tokenID, err := generateUUID()
	if err != nil {
		return err
	}

	tokenSecret, err := generateSecret(s.config.AuthenticationEmailVerificationTokenSecretLength)
	if err != nil {
		return err
	}

	tokenExpiresAt := time.Now().UTC().Add(s.config.AuthenticationEmailVerificationTokenTTL)

	if err := emailVerificationTokenRepo.Create(ctx, tokenID, userID, tokenSecret, tokenExpiresAt); err != nil {
		return err
	}

	// Build a token JWT

	tokenClaims := &EmailVerificationTokenClaims{
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

	// Prepare the verification URL

	verificationURL := s.config.AuthenticationEmailVerificationURL
	verificationURL += "?" + url.Values{"token": []string{tokenString}}.Encode()

	// Render the verification email templates

	var textContentBuf bytes.Buffer
	if err := VerificationEmailTemplateText.Execute(&textContentBuf, map[string]string{
		"VerificationURL": verificationURL,
		"TokenTTL":        s.config.AuthenticationEmailVerificationTokenTTL.String(),
	}); err != nil {
		return err
	}

	var htmlContentBuf bytes.Buffer
	if err := VerificationEmailTemplateHTML.Execute(&htmlContentBuf, map[string]string{
		"VerificationURL": verificationURL,
		"TokenTTL":        s.config.AuthenticationEmailVerificationTokenTTL.String(),
	}); err != nil {
		return err
	}

	// Send the verification email

	if err := s.transactionalEmailService.SendEmail(
		ctx,
		user.Email,
		user.ID,
		"Verify your email address",
		textContentBuf.String(),
		htmlContentBuf.String(),
	); err != nil {
		return err
	}

	// Update the sent_at of the verification code
	if err := emailVerificationTokenRepo.UpdateSentAt(ctx, tokenID, time.Now()); err != nil {
		return err
	}

	// Update the rate limit of the user
	// NOTE: This is not in the same transaction as UpdateSentAt, because it does
	// not have to be 100% correct. Most of the time it should work.
	if err := userRepo.UpdateEmailVerificationRateLimit(
		ctx,
		userID,
		time.Now().UTC().Add(s.config.AuthenticationEmailVerificationRateLimitInterval),
	); err != nil {
		return err
	}

	return nil
}
