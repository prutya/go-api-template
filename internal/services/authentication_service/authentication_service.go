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

var ErrUserRecordLocked = errors.New("user record is locked")
var ErrEmailAlreadyVerified = errors.New("email already verified")
var ErrEmailVerificationCooldown = errors.New("email verification cooldown")
var ErrEmailVerificationExpired = errors.New("email verification expired")
var ErrEmailVerificationNotRequested = errors.New("email verification not requested")
var ErrTooManyOTPAttempts = errors.New("too many OTP attempts")
var ErrInvalidOTP = errors.New("invalid OTP")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidAccessTokenClaims = errors.New("invalid access token claims")
var ErrAccessTokenNotFound = errors.New("access token not found")
var ErrInvalidAccessToken = errors.New("invalid access token")
var ErrInvalidRefreshTokenClaims = errors.New("invalid refresh token claims")
var ErrRefreshTokenNotFound = errors.New("refresh token not found")
var ErrInvalidRefreshToken = errors.New("refresh token invalid")
var ErrRefreshTokenRevoked = errors.New("refresh token revoked")
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionAlreadyTerminated = errors.New("session already terminated")
var ErrSessionExpired = errors.New("session expired")
var ErrPasswordResetCooldown = errors.New("password reset cooldown")
var ErrPasswordResetExpired = errors.New("password reset expired")
var ErrPasswordResetNotRequested = errors.New("password reset not requested")
var ErrInvalidPasswordResetTokenClaims = errors.New("invalid password reset token claims")
var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type PasswordResetTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type AuthenticationService interface {
	Register(ctx context.Context, email string, password string) error
	RequestNewVerificationEmail(ctx context.Context, email string) error
	SendVerificationEmail(ctx context.Context, userID string) error
	// VerifyEmail verifies the email address of a user using the provided token.
	// If successful, logs the user in directly.
	VerifyEmail(
		ctx context.Context,
		email string,
		otp string,
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
	Refresh(ctx context.Context, refreshToken string) (*CreateTokensResult, error)
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
	VerifyPasswordResetOTP(ctx context.Context, email string, otp string) (string, error)
	ResetPassword(
		ctx context.Context,
		passwordResetToken string,
		newPassword string,
		userAgent string,
		ipAddress string,
	) (*CreateTokensResult, error)
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
