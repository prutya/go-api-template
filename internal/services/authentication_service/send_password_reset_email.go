package authentication_service

import (
	"bytes"
	"context"
	text_template "text/template"
	"time"
)

var PasswordResetEmailTemplateText = text_template.Must(text_template.New("password_reset_email").Parse(
	`
	Hi!
	To reset your password, please use the following code:
	{{.Code}}
	This code will expire at {{.CodeExpiresAt}}.
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
		<p><b>{{.Code}}</b></p>
		<p>This code will expire at {{.CodeExpiresAt}}.</p>
		<p>If you did not request a password reset, please ignore this email.</p>
	</body>
	</html>
	`,
))

func (s *authenticationService) SendPasswordResetEmail(ctx context.Context, userID string) error {
	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := findUserByID(ctx, userRepo, userID)
	if err != nil {
		return err
	}

	// An old failed job is retrying but the state has already changed
	if !user.PasswordResetExpiresAt.Valid {
		return ErrPasswordResetNotRequested
	}

	// An old failed job is retrying but the state has already changed
	if user.PasswordResetExpiresAt.Time.Before(time.Now().UTC()) {
		return ErrPasswordResetExpired
	}

	otp, err := generateOtp()
	if err != nil {
		return err
	}

	otpHash, err := generateHmac([]byte(otp), s.config.AuthenticationOtpHmacSecret)
	if err != nil {
		return err
	}

	if err := userRepo.UpdatePasswordResetOtpHmac(ctx, userID, otpHash); err != nil {
		return err
	}

	// Render the email template

	displayCodeExpiresAt := user.PasswordResetExpiresAt.Time.Format(time.RFC3339)

	var textContentBuf bytes.Buffer
	if err := PasswordResetEmailTemplateText.Execute(&textContentBuf, map[string]string{
		"Code":          otp,
		"CodeExpiresAt": displayCodeExpiresAt,
	}); err != nil {
		return err
	}

	var htmlContentBuf bytes.Buffer
	if err := PasswordResetEmailTemplateHTML.Execute(&htmlContentBuf, map[string]string{
		"Code":          otp,
		"CodeExpiresAt": displayCodeExpiresAt,
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

	return nil
}
