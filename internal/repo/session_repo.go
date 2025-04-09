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
	Create(ctx context.Context, sessionID string, userID string) error
	TerminateByID(ctx context.Context, sessionId string, terminatedAt time.Time) error
	TerminateSessionByAccessTokenId(ctx context.Context, accessTokenId string) error
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

func (s *sessionRepo) Create(ctx context.Context, sessionID string, userID string) error {
	session := &models.Session{
		ID:     sessionID,
		UserID: userID,
	}

	if _, err := s.db.NewInsert().Model(session).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *sessionRepo) TerminateByID(ctx context.Context, sessionId string, terminatedAt time.Time) error {
	_, err := s.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("terminated_at = ?", terminatedAt).
		Set("updated_at = now()").
		Where("id = ?", sessionId).
		Where("terminated_at IS NULL").
		Exec(ctx)

	return err
}

func (s *sessionRepo) TerminateSessionByAccessTokenId(ctx context.Context, accessTokenId string) error {
	_, err := s.db.NewUpdate().
		TableExpr("sessions").
		Set("terminated_at = now()").
		Set("updated_at = now()").
		Where(`id IN (
			SELECT sessions.id
			FROM sessions
			JOIN refresh_tokens ON refresh_tokens.session_id = sessions.id
			JOIN access_tokens ON access_tokens.refresh_token_id = refresh_tokens.id
			WHERE access_tokens.id = ?
		)`, accessTokenId).
		Where("terminated_at IS NULL").
		Exec(ctx)

	return err
}
