package transactional_email_service

import (
	"context"
	"time"

	scalewayTransactionalEmails "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
)

type TransactionalEmailService interface {
	SendEmail(ctx context.Context, email string, userID string, subject string, textBody string, htmlBody string) error
	ResetDailyGlobalLimit(ctx context.Context) error
}

type transactionalEmailService struct {
	dailyGlobalLimit int
	senderEmail      string
	senderName       string
	db               bun.IDB
	repoFactory      repo.RepoFactory

	scwTransactionalEmailsAPI *scalewayTransactionalEmails.API
}

func NewTransactionalEmailService(
	ctx context.Context,
	enabled bool,
	dailyGlobalLimit int,
	senderEmail string,
	senderName string,
	scalewayAccessKeyID string,
	scalewaySecretKey string,
	scalewayRegion scw.Region,
	scalewayProjectID string,
	db bun.IDB,
	repoFactory repo.RepoFactory,
) (TransactionalEmailService, error) {
	if !enabled {
		return newNoopTransactionalEmailService(ctx, dailyGlobalLimit, db, repoFactory)
	}

	scwClient, err := scw.NewClient(
		scw.WithAuth(scalewayAccessKeyID, scalewaySecretKey),
		scw.WithDefaultRegion(scalewayRegion),
		scw.WithDefaultProjectID(scalewayProjectID),
	)
	if err != nil {
		return nil, err
	}

	return &transactionalEmailService{
		dailyGlobalLimit: dailyGlobalLimit,
		senderEmail:      senderEmail,
		senderName:       senderName,
		db:               db,
		repoFactory:      repoFactory,

		scwTransactionalEmailsAPI: scalewayTransactionalEmails.NewAPI(scwClient),
	}, nil
}

func (s *transactionalEmailService) SendEmail(
	ctx context.Context,
	email string,
	userID string,
	subject string,
	textBody string,
	htmlBody string,
) error {
	logger := logger.MustFromContext(ctx)

	emailSendAttemptRepo := s.repoFactory.NewEmailSendAttemptRepo(s.db)

	if err := checkGlobalLimit(ctx, s.dailyGlobalLimit, emailSendAttemptRepo, time.Now().UTC()); err != nil {
		return err
	}

	logger.DebugContext(ctx, "Sending transactional email", "subject", subject, "user_id", userID)

	startTime := time.Now()

	_, err := s.scwTransactionalEmailsAPI.CreateEmail(
		&scalewayTransactionalEmails.CreateEmailRequest{
			From: &scalewayTransactionalEmails.CreateEmailRequestAddress{
				Email: s.senderEmail,
				Name:  &s.senderName,
			},
			To: []*scalewayTransactionalEmails.CreateEmailRequestAddress{
				{Email: email},
			},
			Subject: subject,
			Text:    textBody,
			HTML:    htmlBody,
		},
		scw.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	logger.DebugContext(ctx, "Transactional email sent", "duration", duration, "user_id", userID)

	if err := emailSendAttemptRepo.Create(ctx); err != nil {
		return err
	}

	return nil
}

func (s *transactionalEmailService) ResetDailyGlobalLimit(ctx context.Context) error {
	emailSendAttemptRepo := s.repoFactory.NewEmailSendAttemptRepo(s.db)

	return resetDailyGlobalLimit(ctx, emailSendAttemptRepo)
}
