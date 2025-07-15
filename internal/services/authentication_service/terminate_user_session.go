package authentication_service

import (
	"context"
	"time"
)

// NOTE: I am not using transactions here, because it's just a single write
// operation
func (s *authenticationService) TerminateUserSession(
	ctx context.Context,
	accessTokenClaims *AccessTokenClaims,
	sessionID string,
) (bool, error) {
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	// Try to find the session
	session, err := sessionRepo.TryFindByID(ctx, sessionID)
	if err != nil {
		return false, err
	}

	// Make sure that the session belongs to the user
	if session == nil || session.UserID != accessTokenClaims.UserID {
		return false, ErrSessionNotFound
	}

	isCurrentSession := false

	// Find the current session
	currentSession, err := sessionRepo.FindByAccessTokenID(ctx, accessTokenClaims.ID)
	if err != nil {
		return false, err
	}

	// Check if it's the current session
	if session.ID == currentSession.ID {
		isCurrentSession = true
	}

	// Make sure that the sessions is not terminated or expired
	if session.TerminatedAt.Valid {
		return isCurrentSession, ErrSessionAlreadyTerminated
	}

	if session.ExpiresAt.Before(time.Now()) {
		return isCurrentSession, ErrSessionExpired
	}

	// Terminate
	return isCurrentSession, sessionRepo.TerminateByID(ctx, session.ID, time.Now())
}
