package tasks_server

import (
	"context"
	"encoding/json"
	"errors"
	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/transactional_email_service"
	"prutya/go-api-template/internal/tasks"

	"github.com/hibiken/asynq"
)

type sendVerificationEmailTaskHandler struct {
	authenticationService authentication_service.AuthenticationService
}

func newSendVerificationEmailTaskHandler(authenticationService authentication_service.AuthenticationService) *sendVerificationEmailTaskHandler {
	return &sendVerificationEmailTaskHandler{
		authenticationService: authenticationService,
	}
}

func (h *sendVerificationEmailTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendVerificationEmailPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	if err := h.authenticationService.SendVerificationEmail(ctx, payload.UserID); err != nil {
		if errors.Is(err, transactional_email_service.ErrGlobalLimitReached) {
			return asynq.SkipRetry
		}

		return err
	}

	return nil
}
