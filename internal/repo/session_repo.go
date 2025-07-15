package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type SessionRepo interface {
	TryFindByID(ctx context.Context, sessionID string) (*models.Session, error)
	FindByID(ctx context.Context, sessionID string) (*models.Session, error)
	FindByAccessTokenID(ctx context.Context, accessTokenID string) (*models.Session, error)
	Create(
		ctx context.Context,
		sessionID string,
		userID string,
		userAgent string,
		ipAddress string,
		expiresAt time.Time,
	) error
	TerminateByID(ctx context.Context, sessionId string, terminatedAt time.Time) error
	TerminateSessionByAccessTokenId(ctx context.Context, accessTokenId string) error
	TerminateAllSessionsExceptCurrentByUserID(ctx context.Context, userID string, currentSessionID string) error
	TerminateAllSessions(ctx context.Context, userID string) error
	UpdateExpiresAtByID(ctx context.Context, sessionID string, newExpiresAt time.Time) error
	GetActiveForUserWithPagination(
		ctx context.Context,
		userID string,
		pageSize int,
		beforeSession *models.Session,
	) ([]*models.Session, error)
}

type sessionRepo struct {
	db bun.IDB
}

func NewSessionRepo(db bun.IDB) SessionRepo {
	return &sessionRepo{db: db}
}

func (s *sessionRepo) TryFindByID(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.FindByID(ctx, sessionID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return session, err
}

func (s *sessionRepo) FindByID(ctx context.Context, sessionID string) (*models.Session, error) {
	session := &models.Session{ID: sessionID}

	if err := s.db.NewSelect().Model(session).WherePK().Scan(ctx); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionRepo) FindByAccessTokenID(ctx context.Context, accessTokenID string) (*models.Session, error) {
	session := &models.Session{}

	err := s.db.NewSelect().
		Model(session).
		Join("JOIN refresh_tokens rt ON rt.session_id = s.id").
		Join("JOIN access_tokens at ON at.refresh_token_id = rt.id").
		Where("at.id = ?", accessTokenID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionRepo) Create(
	ctx context.Context,
	sessionID string,
	userID string,
	userAgent string,
	ipAddress string,
	expiresAt time.Time,
) error {
	session := &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if userAgent != "" {
		session.UserAgent = sql.NullString{
			String: userAgent,
			Valid:  true,
		}
	}

	if ipAddress != "" {
		session.IPAddress = sql.NullString{
			String: ipAddress,
			Valid:  true,
		}
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
		Where("expires_at > ?", time.Now()).
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
		Where("expires_at > ?", time.Now()).
		Exec(ctx)

	return err
}

func (s *sessionRepo) TerminateAllSessionsExceptCurrentByUserID(ctx context.Context, userID string, currentSessionID string) error {
	_, err := s.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("terminated_at = now()").
		Set("updated_at = now()").
		Where("user_id = ?", userID).
		Where("id != ?", currentSessionID).
		Where("terminated_at IS NULL").
		Where("expires_at > ?", time.Now()).
		Exec(ctx)

	return err
}

func (s *sessionRepo) TerminateAllSessions(ctx context.Context, userID string) error {
	_, err := s.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("terminated_at = now()").
		Set("updated_at = now()").
		Where("user_id = ?", userID).
		Where("terminated_at IS NULL").
		Where("expires_at > ?", time.Now()).
		Exec(ctx)

	return err
}

func (s *sessionRepo) UpdateExpiresAtByID(ctx context.Context, sessionID string, newExpiresAt time.Time) error {
	_, err := s.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("expires_at = ?", newExpiresAt).
		Set("updated_at = now()").
		Where("id = ?", sessionID).
		Exec(ctx)

	return err
}

func (r *sessionRepo) GetActiveForUserWithPagination(
	ctx context.Context,
	userID string,
	pageSize int,
	beforeSession *models.Session,
) ([]*models.Session, error) {
	query := r.db.NewSelect().
		Model(&models.Session{}).
		Where("user_id = ?", userID).
		Where("terminated_at IS NULL").
		Where("expires_at > ?", time.Now()).
		Order("id DESC").
		Limit(pageSize)

	if beforeSession != nil {
		query.Where("id < ?", beforeSession.ID)
	}

	var sessions []*models.Session
	err := query.Scan(ctx, &sessions)

	if errors.Is(err, sql.ErrNoRows) {
		return []*models.Session{}, nil
	}

	return sessions, err
}
