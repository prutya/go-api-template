package repo

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type EmailSendAttemptRepo interface {
	Create(ctx context.Context) error
	CountInRange(ctx context.Context, rangeStart time.Time, rangeEnd time.Time) (int, error)
	DeleteBefore(ctx context.Context, before time.Time) error
}

type emailSendAttemptRepo struct {
	db bun.IDB
}

func NewEmailSendAttemptRepo(db bun.IDB) EmailSendAttemptRepo {
	return &emailSendAttemptRepo{db: db}
}

func (r *emailSendAttemptRepo) Create(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO email_send_attempts DEFAULT VALUES`)
	return err
}

func (r *emailSendAttemptRepo) CountInRange(ctx context.Context, rangeStart time.Time, rangeEnd time.Time) (int, error) {
	return r.db.NewSelect().
		Model((*models.EmailSendAttempt)(nil)).
		Where("attempted_at >= ?", rangeStart).
		Where("attempted_at < ?", rangeEnd).
		Count(ctx)
}

func (r *emailSendAttemptRepo) DeleteBefore(ctx context.Context, before time.Time) error {
	_, err := r.db.NewDelete().
		Model((*models.EmailSendAttempt)(nil)).
		Where("attempted_at < ?", before).
		Exec(ctx)

	return err
}
