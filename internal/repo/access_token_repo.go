// TODO: Test

package repo

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type AccessTokenRepo interface {
	Create(
		ctx context.Context,
		accessTokenId string,
		refreshTokenId string,
		secret []byte,
		expiresAt time.Time,
	) error
	FindById(
		ctx context.Context,
		id string,
	) (*models.AccessToken, error)
}

type accessTokenRepo struct {
	db bun.IDB
}

func NewAccessTokenRepo(db bun.IDB) AccessTokenRepo {
	return &accessTokenRepo{db: db}
}

func (r *accessTokenRepo) Create(
	ctx context.Context,
	accessTokenId string,
	refreshTokenId string,
	secret []byte,
	expiresAt time.Time,
) error {
	accessToken := &models.AccessToken{
		ID:             accessTokenId,
		RefreshTokenID: refreshTokenId,
		Secret:         secret,
		ExpiresAt:      expiresAt,
	}

	if _, err := r.db.NewInsert().Model(accessToken).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *accessTokenRepo) FindById(
	ctx context.Context,
	id string,
) (*models.AccessToken, error) {
	accessToken := &models.AccessToken{}

	err := r.db.NewSelect().Model(accessToken).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}
