package authentication_service

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/services/transactional_email_service"
	"prutya/go-api-template/internal/tasks_client"
)

var ErrEmailIsTaken = errors.New("email already taken")
var ErrUserNotFound = errors.New("user not found")
var ErrEmailAlreadyVerified = errors.New("email already verified")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidAccessTokenClaims = errors.New("invalid access token claims")
var ErrAccessTokenNotFound = errors.New("access token not found")
var ErrInvalidAccessToken = errors.New("invalid token")
var ErrInvalidRefreshTokenClaims = errors.New("invalid refresh token claims")
var ErrRefreshTokenNotFound = errors.New("refresh token not found")
var ErrRefreshTokenInvalid = errors.New("refresh token invalid")
var ErrRefreshTokenRevoked = errors.New("refresh token revoked")
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionAlreadyTerminated = errors.New("session already terminated")
var ErrSessionExpired = errors.New("session expired")
var ErrCSRFTokenMismatch = errors.New("CSRF token mismatch")

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"userId"`
	CSRFToken string `json:"csrfToken"`
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type EmailVerificationTokenClaims struct {
	jwt.RegisteredClaims
}

type PasswordResetTokenClaims struct {
	jwt.RegisteredClaims
}

type AuthenticationService interface {
	Register(ctx context.Context, email string, password string) error
	RequestNewVerificationEmail(ctx context.Context, email string) error
	SendVerificationEmail(ctx context.Context, userID string) error
	// VerifyEmail verifies the email address of a user using the provided token.
	// If successful, logs the user in directly.
	VerifyEmail(
		ctx context.Context,
		emailVerificationToken string,
		userAgent string,
		ipAddress string,
	) (*CreateTokensResult, error)
	CheckIfEmailIsVerified(ctx context.Context, userID string) error
	Login(
		ctx context.Context,
		email string,
		password string,
		userAgent string,
		ipAddress string,
	) (*CreateTokensResult, error)
	Authenticate(ctx context.Context, accessToken string) (*AccessTokenClaims, error)
	Refresh(ctx context.Context, refreshToken string, clientCSRFToken string) (*CreateTokensResult, error)
	Logout(ctx context.Context, accessTokenClaims *AccessTokenClaims) error
	ChangePassword(
		ctx context.Context,
		currentAccessTokenClaims *AccessTokenClaims,
		oldPassword string,
		newPassword string,
		terminateOtherSessions bool,
	) error
	RequestPasswordReset(ctx context.Context, email string) error
	SendPasswordResetEmail(ctx context.Context, userID string) error
	ResetPassword(ctx context.Context, passwordResetToken string, newPassword string) error
	DeleteAccount(ctx context.Context, accessTokenClaims *AccessTokenClaims, password string) error
	GetActiveSessionsForUser(
		ctx context.Context,
		userID string,
		pageSize int,
		beforeCursor *string,
	) (sessions []*models.Session, hasMore bool, err error)
	TerminateUserSession(
		ctx context.Context,
		accessTokenClaims *AccessTokenClaims,
		sessionID string,
	) (hasTerminatedCurrentSession bool, err error)
}

type authenticationService struct {
	config                    *config.Config
	db                        bun.IDB
	repoFactory               repo.RepoFactory
	tasksClient               tasks_client.Client
	transactionalEmailService transactional_email_service.TransactionalEmailService
}

func NewAuthenticationService(
	config *config.Config,
	db bun.IDB,
	repoFactory repo.RepoFactory,
	tasksClient tasks_client.Client,
	transactionalEmailService transactional_email_service.TransactionalEmailService,
) AuthenticationService {
	return &authenticationService{
		config:                    config,
		db:                        db,
		repoFactory:               repoFactory,
		tasksClient:               tasksClient,
		transactionalEmailService: transactionalEmailService,
	}
}
