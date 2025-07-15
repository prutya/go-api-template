package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type PasswordResetToken struct {
	bun.BaseModel `bun:"table:password_reset_tokens,alias:prt"`

	ID        string       `bun:"id,pk"`
	UserID    string       `bun:"user_id"`
	Secret    []byte       `bun:"secret"`
	ExpiresAt time.Time    `bun:"expires_at"`
	SentAt    sql.NullTime `bun:"sent_at"`
	ResetAt   sql.NullTime `bun:"reset_at"`
	CreatedAt time.Time    `bun:"created_at,default:now()"`
	UpdatedAt time.Time    `bun:"updated_at,default:now()"`
}
