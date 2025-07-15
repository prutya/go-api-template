package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type EmailVerificationToken struct {
	bun.BaseModel `bun:"table:email_verification_tokens,alias:evt"`

	ID         string       `bun:"id,pk"`
	UserID     string       `bun:"user_id"`
	Secret     []byte       `bun:"secret"`
	ExpiresAt  time.Time    `bun:"expires_at"`
	SentAt     sql.NullTime `bun:"sent_at"`
	VerifiedAt sql.NullTime `bun:"verified_at"`
	CreatedAt  time.Time    `bun:"created_at,default:now()"`
	UpdatedAt  time.Time    `bun:"updated_at,default:now()"`
}
