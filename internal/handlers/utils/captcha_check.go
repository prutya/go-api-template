package utils

import (
	"net/http"

	"prutya/go-api-template/internal/services/captcha_service"
)

func NewCaptchaCheckMiddleware(captchaService captcha_service.CaptchaService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Read captcha response from headers
			captchaResponse := r.Header.Get("X-Captcha-Response")

			if len(captchaResponse) > 2048 {
				RenderError(w, r, ErrInvalidCaptcha)
				return
			}

			// Verify captcha response
			captchaValid, err := captchaService.Verify(r.Context(), captchaResponse, r.RemoteAddr)
			if err != nil {
				RenderError(w, r, err)
				return
			}

			if !captchaValid {
				RenderError(w, r, ErrInvalidCaptcha)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
