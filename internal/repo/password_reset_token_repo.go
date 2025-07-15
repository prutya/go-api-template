package repo

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type PasswordResetTokenRepo interface {
	Create(ctx context.Context, id string, userID string, secret []byte, expiresAt time.Time) error
	FindByID(ctx context.Context, tokenID string) (*models.PasswordResetToken, error)
	UpdateSentAt(ctx context.Context, id string, sentAt time.Time) error
	MarkAsReset(ctx context.Context, id string) error
}

type passwordResetTokenRepo struct {
	db bun.IDB
}

func NewPasswordResetTokenRepo(db bun.IDB) PasswordResetTokenRepo {
	return &passwordResetTokenRepo{db: db}
}

func (r *passwordResetTokenRepo) Create(
	ctx context.Context,
	id string,
	userID string,
	secret []byte,
	expiresAt time.Time,
) error {
	passwordResetToken := &models.PasswordResetToken{
		ID:        id,
		UserID:    userID,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	if _, err := r.db.NewInsert().Model(passwordResetToken).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *passwordResetTokenRepo) FindByID(ctx context.Context, tokenID string) (*models.PasswordResetToken, error) {
	passwordResetToken := &models.PasswordResetToken{}

	err := r.db.NewSelect().
		Model(passwordResetToken).
		Where("id = ?", tokenID).
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return passwordResetToken, nil
}

func (r *passwordResetTokenRepo) UpdateSentAt(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.PasswordResetToken)(nil)).
		Set("sent_at = ?", sentAt).
		Set("updated_at = now()").
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *passwordResetTokenRepo) MarkAsReset(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().
		Model((*models.PasswordResetToken)(nil)).
		Set("reset_at = now()").
		Set("updated_at = now()").
		Where("id = ?", id).
		Exec(ctx)

	return err
}
