package authentication_service

import (
	"context"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/models"
)

func (s *authenticationService) GetActiveSessionsForUser(
	ctx context.Context,
	userID string,
	pageSize int,
	beforeCursor *string,
) ([]*models.Session, bool, error) {
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	var cursorSession *models.Session

	if beforeCursor != nil {
		cursorActionDb, err := sessionRepo.TryFindByID(ctx, *beforeCursor)

		if err != nil {
			return nil, false, err
		}

		// If the cursor does not exist, do not apply the filter
		if cursorActionDb == nil {
			logger.MustFromContext(ctx).WarnContext(
				ctx,
				"Cursor session does not exist, ignoring filter",
				"user_id", userID,
				"session_id", *beforeCursor,
			)

			beforeCursor = nil
		} else {
			cursorSession = cursorActionDb
		}
	}

	// Get one more item than the page size to determine if there are more items
	sessions, err := sessionRepo.GetActiveForUserWithPagination(ctx, userID, pageSize+1, cursorSession)
	if err != nil {
		return nil, false, err
	}

	// Check if there are more items
	hasMore := false

	if len(sessions) > pageSize {
		hasMore = true
		sessions = sessions[:pageSize]
	}

	return sessions, hasMore, nil
}
