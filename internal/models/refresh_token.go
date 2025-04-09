package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type RefreshToken struct {
	bun.BaseModel `bun:"table:refresh_tokens,alias:rt"`

	ID              string         `bun:"id,pk"`
	SessionID       string         `bun:"session_id"`
	ParentID        sql.NullString `bun:"parent_id"`
	Secret          []byte         `bun:"secret"`
	ExpiresAt       time.Time      `bun:"expires_at"`
	RevokedAt       sql.NullTime   `bun:"revoked_at"`
	LeewayExpiresAt sql.NullTime   `bun:"leeway_expires_at"`
	CreatedAt       time.Time      `bun:"created_at,default:now()"`
	UpdatedAt       time.Time      `bun:"updated_at,default:now()"`
}
