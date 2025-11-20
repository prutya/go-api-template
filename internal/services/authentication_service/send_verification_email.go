package authentication_service

import (
	"bytes"
	"context"
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

func (s *authenticationService) SendVerificationEmail(ctx context.Context, userID string) error {
	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := findUserByID(ctx, userRepo, userID)
	if err != nil {
		return err
	}

	// An old failed job is retrying but the state has already changed
	if user.EmailVerifiedAt.Valid {
		return ErrEmailAlreadyVerified
	}

	// An old failed job is retrying but the state has already changed
	if !user.EmailVerificationExpiresAt.Valid {
		return ErrEmailVerificationNotRequested
	}

	// An old failed job is retrying but the state has already changed
	if user.EmailVerificationExpiresAt.Time.Before(time.Now().UTC()) {
		return ErrEmailVerificationExpired
	}

	otp, err := generateOtp()
	if err != nil {
		return err
	}

	optHash, err := s.argon2GenerateHashFromOTP(otp)
	if err != nil {
		return err
	}

	if err := userRepo.UpdateEmailVerificationOtpDigest(ctx, userID, optHash); err != nil {
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
