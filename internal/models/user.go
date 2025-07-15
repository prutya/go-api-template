package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                                string       `bun:"id,pk"`
	Email                             string       `bun:"email"`
	PasswordDigest                    string       `bun:"password_digest"`
	EmailVerificationRateLimitedUntil sql.NullTime `bun:"email_verification_rate_limited_until"`
	EmailVerifiedAt                   sql.NullTime `bun:"email_verified_at"`
	PasswordResetRateLimitedUntil     sql.NullTime `bun:"password_reset_rate_limited_until"`
	CreatedAt                         time.Time    `bun:"created_at,default:now()"`
	UpdatedAt                         time.Time    `bun:"updated_at,default:now()"`
}
