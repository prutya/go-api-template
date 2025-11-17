package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type RefreshTokenRepo interface {
	Create(
		ctx context.Context,
		refreshTokenId string,
		sessionId string,
		parentId sql.NullString,
		secret []byte,
		expiresAt time.Time,
	) error
	FindById(ctx context.Context, id string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time, leewayExpiresAt time.Time) error
}

type refreshTokenRepo struct {
	db bun.IDB
}

func NewRefreshTokenRepo(db bun.IDB) RefreshTokenRepo {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(
	ctx context.Context,
	refreshTokenId string,
	sessionId string,
	parentId sql.NullString,
	secret []byte,
	expiresAt time.Time,
) error {
	refreshToken := &models.RefreshToken{
		ID:        refreshTokenId,
		SessionID: sessionId,
		ParentID:  parentId,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	if _, err := r.db.NewInsert().Model(refreshToken).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *refreshTokenRepo) FindById(
	ctx context.Context,
	id string,
) (*models.RefreshToken, error) {
	refreshToken := &models.RefreshToken{}

	err := r.db.NewSelect().Model(refreshToken).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (r *refreshTokenRepo) Revoke(
	ctx context.Context,
	id string,
	revokedAt time.Time,
	leewayExpiresAt time.Time,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.RefreshToken)(nil)).
		Set("revoked_at = ?", revokedAt).
		Set("leeway_expires_at = ?", leewayExpiresAt).
		Set("updated_at = now()").
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
