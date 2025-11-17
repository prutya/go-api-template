package transactional_email_service

import (
	"context"
	"log/slog"
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

func newNoopTransactionalEmailService(
	dailyGlobalLimit int,
	db bun.IDB,
	repoFactory repo.RepoFactory,
) (TransactionalEmailService, error) {
	slog.Warn("Transactional emails are disabled. Delivery is faked")

	return &noopTransactionalEmailService{
		dailyGlobalLimit: dailyGlobalLimit,
		db:               db,
		repoFactory:      repoFactory,
	}, nil
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
		"Fake transactional email",
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
