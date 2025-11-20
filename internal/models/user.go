package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             string `bun:"id,pk"`
	Email          string `bun:"email"`
	PasswordDigest string `bun:"password_digest"`

	EmailVerifiedAt                   sql.NullTime `bun:"email_verified_at"`
	EmailVerificationOtpDigest        string       `bun:"email_verification_otp_digest"`
	EmailVerificationExpiresAt        sql.NullTime `bun:"email_verification_expires_at"`
	EmailVerificationOtpAttempts      int          `bun:"email_verification_otp_attempts"`
	EmailVerificationCooldownResetsAt sql.NullTime `bun:"email_verification_cooldown_resets_at"`
	EmailVerificationLastRequestedAt  sql.NullTime `bun:"email_verification_last_requested_at"`

	PasswordResetOtpDigest        string       `bun:"password_reset_otp_digest"`
	PasswordResetExpiresAt        sql.NullTime `bun:"password_reset_expires_at"`
	PasswordResetOtpAttempts      int          `bun:"password_reset_otp_attempts"`
	PasswordResetCooldownResetsAt sql.NullTime `bun:"password_reset_cooldown_resets_at"`
	PasswordResetLastRequestedAt  sql.NullTime `bun:"password_reset_last_requested_at"`
	PasswordResetTokenPublicKey   []byte       `bun:"password_reset_token_public_key"`

	CreatedAt time.Time `bun:"created_at,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,default:now()"`
}
