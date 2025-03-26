// TODO: Tests

package authentication_service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
	"prutya/go-api-template/internal/repo"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidTokenClaims = errors.New("invalid token claims")
var ErrSessionNotFound = errors.New("session not found")
var ErrInvalidToken = errors.New("invalid token")
var ErrSessionExpired = errors.New("session expired")
var ErrSessionTerminated = errors.New("session terminated")
var ErrUserNotFound = errors.New("user not found")

type SessionTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userId"`
}

type AuthenticationService interface {
	Login(ctx context.Context, email string, password string) (sessionToken string, err error)
	Logout(ctx context.Context, sessionId string) error
	Authenticate(ctx context.Context, sessionToken string) (user *models.User, session *models.Session, err error)
}

type authenticationService struct {
	config      *config.Config
	userRepo    repo.UserRepo
	sessionRepo repo.SessionRepo
}

func NewAuthenticationService(config *config.Config, userRepo repo.UserRepo, sessionRepo repo.SessionRepo) AuthenticationService {
	return &authenticationService{
		config:      config,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *authenticationService) Login(ctx context.Context, email string, password string) (string, error) {
	normalizedEmail := strings.ToLower(email)

	user, err := s.userRepo.FindByEmail(ctx, normalizedEmail)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}

		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidCredentials
		}

		return "", err
	}

	sessionIdUUID, err := uuid.NewV7()

	if err != nil {
		return "", err
	}

	sessionId := sessionIdUUID.String()

	sessionSecret := make([]byte, s.config.AuthenticationSessionTokenSecretLength)

	_, err = rand.Read(sessionSecret)

	if err != nil {
		return "", err
	}

	sessionExpiresAt := time.Now().Add(s.config.AuthenticationSessionTokenTTL)

	err = s.sessionRepo.Create(ctx, user.ID, sessionId, sessionSecret, sessionExpiresAt)

	if err != nil {
		return "", err
	}

	claims := SessionTokenClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionId,
			ExpiresAt: jwt.NewNumericDate(sessionExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(sessionSecret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authenticationService) Logout(ctx context.Context, sessionId string) error {
	return s.sessionRepo.Terminate(ctx, sessionId)
}

func (s *authenticationService) Authenticate(
	ctx context.Context,
	sessionToken string,
) (*models.User, *models.Session, error) {
	logger := logger.MustFromContext(ctx)

	var userSession *models.Session

	// Parse the token
	parsedToken, err := jwt.ParseWithClaims(sessionToken, &SessionTokenClaims{}, func(token *jwt.Token) (any, error) {
		// Find the session by JTI
		claims, ok := token.Claims.(*SessionTokenClaims)
		if !ok {
			return nil, ErrInvalidTokenClaims
		}

		session, err := s.sessionRepo.FindById(ctx, claims.ID)
		if err != nil {
			logger.Warn("Session not found", zap.String("session_id", claims.ID))

			return nil, ErrSessionNotFound
		}

		userSession = session

		return session.Secret, nil
	})

	if err != nil {
		if errors.Is(err, ErrInvalidTokenClaims) || errors.Is(err, ErrSessionNotFound) {
			return nil, nil, err
		}

		return nil, nil, ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(*SessionTokenClaims)
	if !ok {
		return nil, nil, ErrInvalidTokenClaims
	}

	// Check if session is expired
	if userSession.ExpiresAt.Before(time.Now()) {
		logger.Warn("Session expired", zap.String("session_id", userSession.ID))

		return nil, nil, ErrSessionExpired
	}

	// Check if session is terminated
	if userSession.TerminatedAt.Valid {
		logger.Warn("Session terminated", zap.String("session_id", userSession.ID))

		return nil, nil, ErrSessionTerminated
	}

	// Find the user
	user, err := s.userRepo.FindById(ctx, claims.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	return user, userSession, nil
}
