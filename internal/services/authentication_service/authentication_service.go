// TODO: Tests

package authentication_service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/tasks_client"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidAccessTokenClaims = errors.New("invalid access token claims")
var ErrAccessTokenNotFound = errors.New("access token not found")
var ErrInvalidAccessToken = errors.New("invalid token")
var ErrInvalidRefreshTokenClaims = errors.New("invalid refresh token claims")
var ErrRefreshTokenNotFound = errors.New("refresh token not found")
var ErrRefreshTokenInvalid = errors.New("refresh token invalid")
var ErrRefreshTokenRevoked = errors.New("refresh token revoked")
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionTerminated = errors.New("session terminated")

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type AuthenticationService interface {
	Login(ctx context.Context, email string, password string) (refreshToken string, refreshTokenExpiresAt time.Time, accessToken string, err error)
	Authenticate(ctx context.Context, accessToken string) (*AccessTokenClaims, error)
	Refresh(ctx context.Context, refreshToken string) (newRefreshToken string, newRefreshTokenExpiresAt time.Time, newAccessToken string, err error)
	Logout(ctx context.Context, accessTokenClaims *AccessTokenClaims) error
}

type authenticationService struct {
	config           *config.Config
	userRepo         repo.UserRepo
	sessionRepo      repo.SessionRepo
	refreshTokenRepo repo.RefreshTokenRepo
	accessTokenRepo  repo.AccessTokenRepo
	tasksClient      tasks_client.Client
}

func NewAuthenticationService(
	config *config.Config,
	userRepo repo.UserRepo,
	sessionRepo repo.SessionRepo,
	refreshTokenRepo repo.RefreshTokenRepo,
	accessTokenRepo repo.AccessTokenRepo,
	tasksClient tasks_client.Client,
) AuthenticationService {
	return &authenticationService{
		config:           config,
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		refreshTokenRepo: refreshTokenRepo,
		accessTokenRepo:  accessTokenRepo,
		tasksClient:      tasksClient,
	}
}
