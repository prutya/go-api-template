package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID           string         `bun:"id,pk"`
	UserID       string         `bun:"user_id"`
	UserAgent    sql.NullString `bun:"user_agent"`
	IPAddress    sql.NullString `bun:"ip_address"`
	TerminatedAt sql.NullTime   `bun:"terminated_at"`
	ExpiresAt    time.Time      `bun:"expires_at"`
	CreatedAt    time.Time      `bun:"created_at,default:now()"`
	UpdatedAt    time.Time      `bun:"updated_at,default:now()"`
}
