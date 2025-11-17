package authentication_service

import (
	"bytes"
	"context"
	"errors"
	html_template "html/template"
	text_template "text/template"
	"time"
)

var VerificationEmailTemplateText = text_template.Must(text_template.New("verification_email").Parse(
	`
	Hi!
	Thank you for signing up! To complete your registration, please use the following code:
	{{.Code}}
	This code will expire at {{.CodeExpiresAt}}.
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
		<p>Thank you for signing up! To complete your registration, please use the following code:</p>
		<p><b>{{.Code}}</b></p>
		<p>This code will expire at {{.CodeExpiresAt}}.</p>
		<p>If you did not sign up for this account, please ignore this email.</p>
	</body>
	</html>
	`,
))

var ErrEmailVerificationRateLimited = errors.New("email verification rate limited")

func (s *authenticationService) SendVerificationEmail(ctx context.Context, userID string) error {
	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := findUserByID(ctx, userRepo, userID)
	if err != nil {
		return err
	}

	// TODO: If email verification has already expired, throw an error

	otp, err := generateOtp()
	if err != nil {
		return err
	}

	otpHash, err := generateHmac([]byte(otp), s.config.AuthenticationOtpHmacSecret)
	if err != nil {
		return err
	}

	if err := userRepo.UpdateEmailVerificationOtpHmac(ctx, userID, otpHash); err != nil {
		return err
	}

	// Render the verification email templates

	var textContentBuf bytes.Buffer
	if err := VerificationEmailTemplateText.Execute(&textContentBuf, map[string]string{
		"Code":          otp,
		"CodeExpiresAt": user.EmailVerificationExpiresAt.Time.Format(time.RFC3339),
	}); err != nil {
		return err
	}

	var htmlContentBuf bytes.Buffer
	if err := VerificationEmailTemplateHTML.Execute(&htmlContentBuf, map[string]string{
		"Code":          otp,
		"CodeExpiresAt": user.EmailVerificationExpiresAt.Time.Format(time.RFC3339),
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

	return nil
}
