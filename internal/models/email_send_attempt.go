package models

import (
	"time"

	"github.com/uptrace/bun"
)

type EmailSendAttempt struct {
	bun.BaseModel `bun:"table:email_send_attempts,alias:esa"`

	ID          int       `bun:"id,pk"`
	AttemptedAt time.Time `bun:"attempted_at,default:now()"`
}
