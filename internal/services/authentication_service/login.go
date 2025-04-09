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
	"golang.org/x/crypto/bcrypt"
)

func (s *authenticationService) Login(ctx context.Context, email string, password string) (string, time.Time, string, error) {
	// Prevent the potential attacker from measuring the response time
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		timeLeft := s.config.AuthenticationTimingAttackDelay - duration

		if timeLeft > 0 {
			time.Sleep(timeLeft)
		}
	}()

	normalizedEmail := strings.ToLower(email)

	// Find the user by email

	user, err := s.userRepo.FindByEmail(ctx, normalizedEmail)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, "", ErrInvalidCredentials
		}

		return "", time.Time{}, "", err
	}

	// Check if the password is correct

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", time.Time{}, "", ErrInvalidCredentials
		}

		return "", time.Time{}, "", err
	}

	// Create a session

	sessionIdUUID, err := uuid.NewV7()

	if err != nil {
		return "", time.Time{}, "", err
	}

	sessionId := sessionIdUUID.String()

	err = s.sessionRepo.Create(ctx, sessionId, user.ID)

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Create a RefreshToken

	refreshTokenUUID, err := uuid.NewV7()

	if err != nil {
		return "", time.Time{}, "", err
	}

	refreshTokenId := refreshTokenUUID.String()

	refreshTokenSecret := make([]byte, s.config.AuthenticationRefreshTokenSecretLength)

	_, err = rand.Read(refreshTokenSecret)

	if err != nil {
		return "", time.Time{}, "", err
	}

	refreshTokenExpiresAt := time.Now().Add(s.config.AuthenticationRefreshTokenTTL)

	if err := s.refreshTokenRepo.Create(
		ctx,
		refreshTokenId,
		sessionId,
		sql.NullString{},
		refreshTokenSecret,
		refreshTokenExpiresAt,
	); err != nil {
		return "", time.Time{}, "", err
	}

	// Create an AccessToken

	accessTokenUUID, err := uuid.NewV7()

	if err != nil {
		return "", time.Time{}, "", err
	}

	accessTokenId := accessTokenUUID.String()

	accessTokenSecret := make([]byte, s.config.AuthenticationAccessTokenSecretLength)

	_, err = rand.Read(accessTokenSecret)
	if err != nil {
		return "", time.Time{}, "", err
	}

	accessTokenExpiresAt := time.Now().Add(s.config.AuthenticationAccessTokenTTL)

	if err := s.accessTokenRepo.Create(
		ctx,
		accessTokenId,
		refreshTokenId,
		accessTokenSecret,
		accessTokenExpiresAt,
	); err != nil {
		return "", time.Time{}, "", err
	}

	// Create a JWT for the refresh token
	refreshTokenClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenId,
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: user.ID,
	}

	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	refreshTokenString, err := refreshTokenJWT.SignedString(refreshTokenSecret)

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Create a JWT for the access token

	accessTokenClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenId,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: user.ID,
	}

	accessTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	accessTokenString, err := accessTokenJWT.SignedString(accessTokenSecret)

	if err != nil {
		return "", time.Time{}, "", err
	}

	return refreshTokenString, refreshTokenExpiresAt, accessTokenString, nil
}
