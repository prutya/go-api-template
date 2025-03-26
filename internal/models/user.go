package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             string    `bun:"id,pk"`
	Email          string    `bun:"email"`
	PasswordDigest string    `bun:"password_digest"`
	CreatedAt      time.Time `bun:"created_at,default:now()"`
	UpdatedAt      time.Time `bun:"updated_at,default:now()"`
}
