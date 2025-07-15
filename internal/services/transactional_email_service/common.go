package transactional_email_service

import (
	"context"
	"errors"
	"time"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
)

var ErrGlobalLimitReached = errors.New("global limit reached")

func checkGlobalLimit(
	ctx context.Context,
	dailyGlobalLimit int,
	emailSendAttemptRepo repo.EmailSendAttemptRepo,
	currentTime time.Time,
) error {
	logger := logger.MustFromContext(ctx)

	if dailyGlobalLimit <= 0 {
		logger.WarnContext(ctx, "Daily global email limit is <= 0, no emails will be sent")

		return ErrGlobalLimitReached
	}

	rangeStart := currentTime.Truncate(24 * time.Hour)
	rangeEnd := rangeStart.Add(24 * time.Hour)

	currentCount, err := emailSendAttemptRepo.CountInRange(ctx, rangeStart, rangeEnd)
	if err != nil {
		return err
	}

	if currentCount >= dailyGlobalLimit {
		logger.WarnContext(
			ctx,
			"Daily global email limit reached, no emails will be sent",
			"current_count", currentCount,
			"daily_global_limit", dailyGlobalLimit,
		)

		return ErrGlobalLimitReached
	}

	return nil
}

func resetDailyGlobalLimit(ctx context.Context, emailSendAttemptRepo repo.EmailSendAttemptRepo) error {
	startOfDay := time.Now().Truncate(24 * time.Hour)

	if err := emailSendAttemptRepo.DeleteBefore(ctx, startOfDay); err != nil {
		return err
	}

	logger.MustFromContext(ctx).InfoContext(ctx, "Daily global email limit reset")

	return nil
}
