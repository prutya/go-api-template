package tasks_server

import (
	"context"
	"prutya/go-api-template/internal/services/transactional_email_service"

	"github.com/hibiken/asynq"
)

type cleanupEmailSendAttemptsHandler struct {
	transactionalEmailService transactional_email_service.TransactionalEmailService
}

func newCleanupEmailSendAttemptsHandler(
	transactionalEmailService transactional_email_service.TransactionalEmailService,
) *cleanupEmailSendAttemptsHandler {
	return &cleanupEmailSendAttemptsHandler{
		transactionalEmailService: transactionalEmailService,
	}
}

func (h *cleanupEmailSendAttemptsHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return h.transactionalEmailService.ResetDailyGlobalLimit(ctx)
}
