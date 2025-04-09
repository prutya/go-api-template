package user_service

import (
	"context"
	"database/sql"
	"errors"

	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/repo"
)

type UserService interface {
	GetUserById(ctx context.Context, id string) (*models.User, error)
}

var ErrUserNotFound = errors.New("user not found")

type userService struct {
	userRepo repo.UserRepo
}

func NewUserService(userRepo repo.UserRepo) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserById(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.FindById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, err
}
