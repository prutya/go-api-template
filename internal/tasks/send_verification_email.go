package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeSendVerificationEmail = "send_verification_email"

type SendVerificationEmailPayload struct {
	UserID string
}

func NewSendVerificationEmailTask(userID string) (*Task, error) {
	payload, err := json.Marshal(SendVerificationEmailPayload{
		UserID: userID,
	})

	if err != nil {
		return nil, err
	}

	return NewTask(asynq.NewTask(TypeSendVerificationEmail, payload)), nil
}
