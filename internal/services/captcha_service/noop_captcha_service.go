package captcha_service

import (
	"context"
	"log/slog"
)

type noopCaptchaService struct{}

func newNoopCaptchaService() CaptchaService {
	slog.Warn("Captcha verification is disabled. All verification attempts will succeed.")

	return &noopCaptchaService{}
}

func (s *noopCaptchaService) Verify(ctx context.Context, captchaResponse string, ip string) (bool, error) {
	return true, nil
}
