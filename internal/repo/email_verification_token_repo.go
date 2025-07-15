package repo

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type EmailVerificationTokenRepo interface {
	Create(ctx context.Context, id string, userID string, secret []byte, expiresAt time.Time) error
	FindByID(ctx context.Context, tokenID string) (*models.EmailVerificationToken, error)
	UpdateSentAt(ctx context.Context, id string, sentAt time.Time) error
	MarkAsVerified(ctx context.Context, id string) error
}

type emailVerificationTokenRepo struct {
	db bun.IDB
}

func NewEmailVerificationTokenRepo(db bun.IDB) EmailVerificationTokenRepo {
	return &emailVerificationTokenRepo{db: db}
}

func (r *emailVerificationTokenRepo) Create(
	ctx context.Context,
	id string,
	userID string,
	secret []byte,
	expiresAt time.Time,
) error {
	emailVerificationToken := &models.EmailVerificationToken{
		ID:        id,
		UserID:    userID,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	if _, err := r.db.NewInsert().Model(emailVerificationToken).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *emailVerificationTokenRepo) FindByID(ctx context.Context, tokenID string) (*models.EmailVerificationToken, error) {
	emailVerificationToken := &models.EmailVerificationToken{}

	err := r.db.NewSelect().
		Model(emailVerificationToken).
		Where("id = ?", tokenID).
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return emailVerificationToken, nil
}

func (r *emailVerificationTokenRepo) UpdateSentAt(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.EmailVerificationToken)(nil)).
		Where("id = ?", id).
		Set("sent_at = ?", sentAt).
		Set("updated_at = now()").
		Exec(ctx)

	return err
}

func (r *emailVerificationTokenRepo) MarkAsVerified(ctx context.Context, codeID string) error {
	_, err := r.db.NewUpdate().
		Model((*models.EmailVerificationToken)(nil)).
		Where("id = ?", codeID).
		Set("verified_at = now()").
		Set("updated_at = now()").
		Exec(ctx)

	return err
}
