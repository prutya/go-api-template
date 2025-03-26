// TODO: Tests

package repo

import (
	"context"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type UserRepo interface {
	FindById(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepo struct {
	db bun.IDB
}

func NewUserRepo(db bun.IDB) UserRepo {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) FindById(ctx context.Context, id string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}
