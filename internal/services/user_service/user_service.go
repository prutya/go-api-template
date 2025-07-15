package user_service

import (
	"context"
	"database/sql"
	"errors"

	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/repo"

	"github.com/uptrace/bun"
)

var ErrUserNotFound = errors.New("user not found")

type UserService interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type userService struct {
	db          bun.IDB
	repoFactory repo.RepoFactory
}

func NewUserService(db bun.IDB, repoFactory repo.RepoFactory) UserService {
	return &userService{
		db:          db,
		repoFactory: repoFactory,
	}
}

func (s *userService) FindByID(ctx context.Context, id string) (*models.User, error) {
	userRepo := s.repoFactory.NewUserRepo(s.db)

	user, err := userRepo.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}
