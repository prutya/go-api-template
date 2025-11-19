package captcha_service

import (
	"context"
	"prutya/go-api-template/internal/logger"
)

type noopCaptchaService struct{}

func newNoopCaptchaService(ctx context.Context) CaptchaService {
	logger.MustWarnContext(ctx, "Captcha verification is disabled. All verification attempts will succeed.")

	return &noopCaptchaService{}
}

func (s *noopCaptchaService) Verify(ctx context.Context, captchaResponse string, ip string) (bool, error) {
	return true, nil
}
