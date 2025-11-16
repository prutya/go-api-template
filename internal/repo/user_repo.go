package repo

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type UserRepo interface {
	Create(ctx context.Context, id string, email string, password string) error
	UpdateEmailVerificationRateLimit(ctx context.Context, userID string, rateLimitUntil time.Time) error
	UpdatePasswordResetRateLimit(ctx context.Context, userID string, rateLimitUntil time.Time) error
	MarkEmailAsVerified(ctx context.Context, userID string) error
	FindByID(ctx context.Context, userID string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	UpdatePasswordDigest(ctx context.Context, userID string, newPasswordDigest string) error
	Delete(ctx context.Context, userID string) error
}

type userRepo struct {
	db bun.IDB
}

func NewUserRepo(db bun.IDB) UserRepo {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Create(ctx context.Context, id string, email string, passwordDigest string) error {
	user := &models.User{
		ID:             id,
		Email:          email,
		PasswordDigest: passwordDigest,
	}

	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) UpdateEmailVerificationRateLimit(ctx context.Context, userID string, rateLimitUntil time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Where("id = ?", userID).
		Set("email_verification_rate_limited_until = ?", rateLimitUntil).
		Set("updated_at = now()").
		Exec(ctx)

	return err
}

func (r *userRepo) UpdatePasswordResetRateLimit(ctx context.Context, userID string, rateLimitUntil time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Where("id = ?", userID).
		Set("password_reset_rate_limited_until = ?", rateLimitUntil).
		Set("updated_at = now()").
		Exec(ctx)

	return err
}

func (r *userRepo) MarkEmailAsVerified(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Where("id = ?", userID).
		Set("email_verified_at = now()").
		Set("updated_at = now()").
		Exec(ctx)

	return err
}

func (r *userRepo) FindByID(ctx context.Context, userID string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", userID).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("lower(email) = lower(?)", email).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) UpdatePasswordDigest(ctx context.Context, userID string, newPasswordDigest string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_digest = ?", newPasswordDigest).
		Set("updated_at = now()").
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().
		Model((*models.User)(nil)).
		Where("id = ?", userID).
		Exec(ctx)

	return err
}
