// TODO: Tests

package repo

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type SessionRepo interface {
	FindById(ctx context.Context, sessionID string) (*models.Session, error)
	Create(ctx context.Context, userID string, sessionID string, secret []byte, expiresAt time.Time) error
	Terminate(ctx context.Context, sessionID string) error
}

type sessionRepo struct {
	db bun.IDB
}

func NewSessionRepo(db bun.IDB) SessionRepo {
	return &sessionRepo{db: db}
}

func (s *sessionRepo) FindById(ctx context.Context, sessionID string) (*models.Session, error) {
	session := &models.Session{ID: sessionID}

	if err := s.db.NewSelect().Model(session).WherePK().Scan(ctx); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionRepo) Create(ctx context.Context, userID string, sessionID string, secret []byte, expiresAt time.Time) error {
	session := &models.Session{
		ID:        sessionID,
		UserID:    userID,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	if _, err := s.db.NewInsert().Model(session).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *sessionRepo) Terminate(ctx context.Context, sessionID string) error {
	session := &models.Session{ID: sessionID}

	// Set the terminated_at to the current time
	if _, err := s.db.NewUpdate().Model(session).Set("terminated_at = now()").WherePK().Exec(ctx); err != nil {
		return err
	}

	return nil
}
