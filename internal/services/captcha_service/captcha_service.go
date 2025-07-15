package captcha_service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"prutya/go-api-template/internal/logger"
	"time"
)

type CaptchaService interface {
	Verify(ctx context.Context, captchaResponse string, ip string) (bool, error)
}

type captchaService struct {
	httpClient *http.Client
	baseURL    string
	secretKey  string
}

func NewCaptchaService(
	enabled bool,
	baseURL string,
	secretKey string,
) CaptchaService {
	if !enabled {
		return &noopCaptchaService{}
	}

	// TODO: Make configurable
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxConnsPerHost = 100
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.IdleConnTimeout = 90 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second

	// TODO: Make configurable
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Maximum redirects
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}

			return nil
		},
	}

	return &captchaService{
		httpClient: httpClient,
		baseURL:    baseURL,
		secretKey:  secretKey,
	}
}

type turnstileRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip"`
}

// NOTE: The actual response from the Turnstile API has more fields
type turnstilePartialResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes,omitempty"`
}

func (s *captchaService) Verify(ctx context.Context, captchaResponse string, ip string) (bool, error) {
	logger := logger.MustFromContext(ctx)

	requestBody := &turnstileRequest{
		Secret:   s.secretKey,
		Response: captchaResponse,
		RemoteIP: ip,
	}

	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, err
	}

	logger.DebugContext(ctx, "sending captcha verification request")

	startTime := time.Now()

	response, err := s.httpClient.Post(
		s.baseURL+"/siteverify",
		"application/json",
		bytes.NewBuffer(jsonRequestBody),
	)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	requestDuration := time.Since(startTime)

	logger.DebugContext(ctx, "captcha verification response", "status_code", response.StatusCode, "duration", requestDuration)

	responseBody := &turnstilePartialResponse{}
	if err := json.NewDecoder(response.Body).Decode(responseBody); err != nil {
		return false, err
	}

	if !responseBody.Success {
		logger.WarnContext(ctx, "captcha validation failed", "error_codes", responseBody.ErrorCodes)

		return false, nil
	}

	logger.DebugContext(ctx, "captcha validation succeeded")

	return true, nil
}
