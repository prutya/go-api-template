package models

import (
	"time"

	"github.com/uptrace/bun"
)

type AccessToken struct {
	bun.BaseModel `bun:"table:access_tokens,alias:at"`

	ID             string    `bun:"id,pk"`
	RefreshTokenID string    `bun:"refresh_token_id"`
	Secret         []byte    `bun:"secret"`
	ExpiresAt      time.Time `bun:"expires_at"`
	CreatedAt      time.Time `bun:"created_at,default:now()"`
	UpdatedAt      time.Time `bun:"updated_at,default:now()"`
}
