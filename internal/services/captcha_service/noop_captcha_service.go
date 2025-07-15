package captcha_service

import (
	"context"
	"prutya/go-api-template/internal/logger"
)

type noopCaptchaService struct{}

func (s *noopCaptchaService) Verify(ctx context.Context, captchaResponse string, ip string) (bool, error) {
	logger.MustFromContext(ctx).WarnContext(ctx, "Captcha verification is disabled, always returning true")

	return true, nil
}
