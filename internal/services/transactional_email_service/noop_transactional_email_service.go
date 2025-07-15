package transactional_email_service

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
)

type noopTransactionalEmailService struct {
	dailyGlobalLimit int
	db               bun.IDB
	repoFactory      repo.RepoFactory
}

func (s *noopTransactionalEmailService) SendEmail(
	ctx context.Context,
	email string,
	userID string,
	subject string,
	textBody string,
	_ string,
) error {
	logger := logger.MustFromContext(ctx)

	emailSendAttemptRepo := s.repoFactory.NewEmailSendAttemptRepo(s.db)

	if err := checkGlobalLimit(ctx, s.dailyGlobalLimit, emailSendAttemptRepo, time.Now()); err != nil {
		return err
	}

	logger.WarnContext(
		ctx,
		"Transactional email sending is disabled, faking email sending",
		"subject", subject,
		"text_body", textBody,
		"user_id", userID,
	)

	if err := emailSendAttemptRepo.Create(ctx); err != nil {
		return err
	}

	return nil
}

func (s *noopTransactionalEmailService) ResetDailyGlobalLimit(ctx context.Context) error {
	emailSendAttemptRepo := s.repoFactory.NewEmailSendAttemptRepo(s.db)

	return resetDailyGlobalLimit(ctx, emailSendAttemptRepo)
}
