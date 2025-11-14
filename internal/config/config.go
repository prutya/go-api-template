package config

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel             string        `mapstructure:"LOG_LEVEL"`
	LogFormat            string        `mapstructure:"LOG_FORMAT"`
	LogTimeFormat        string        `mapstructure:"LOG_TIME_FORMAT"`
	RequestTimeout       time.Duration `mapstructure:"REQUEST_TIMEOUT"`
	CorsAllowedOrigins   []string      `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CorsAllowedMethods   []string      `mapstructure:"CORS_ALLOWED_METHODS"`
	CorsAllowedHeaders   []string      `mapstructure:"CORS_ALLOWED_HEADERS"`
	CorsExposedHeaders   []string      `mapstructure:"CORS_EXPOSED_HEADERS"`
	CorsAllowCredentials bool          `mapstructure:"CORS_ALLOW_CREDENTIALS"`
	CorsMaxAge           time.Duration `mapstructure:"CORS_MAX_AGE"`
	ShutdownTimeout      time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
	ListenAddr           string        `mapstructure:"LISTEN_ADDR"`
	ReadTimeout          time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout         time.Duration `mapstructure:"WRITE_TIMEOUT"`
	ReadHeaderTimeout    time.Duration `mapstructure:"READ_HEADER_TIMEOUT"`
	IdleTimeout          time.Duration `mapstructure:"IDLE_TIMEOUT"`

	DatabaseUrl             string        `mapstructure:"DATABASE_URL"`
	DatabaseMaxOpenConns    int           `mapstructure:"DATABASE_MAX_OPEN_CONNS"`
	DatabaseMaxIdleConns    int           `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
	DatabaseMaxConnLifetime time.Duration `mapstructure:"DATABASE_MAX_CONN_LIFETIME"`
	DatabaseMaxConnIdleTime time.Duration `mapstructure:"DATABASE_MAX_CONN_IDLE_TIME"`

	AuthenticationBcryptCost                         int           `mapstructure:"AUTHENTICATION_BCRYPT_COST"`
	AuthenticationTimingAttackDelay                  time.Duration `mapstructure:"AUTHENTICATION_TIMING_ATTACK_DELAY"`
	AuthenticationRefreshTokenTTL                    time.Duration `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_TTL"`
	AuthenticationRefreshTokenSecretLength           int           `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_SECRET_LENGTH"`
	AuthenticationRefreshTokenLeeway                 time.Duration `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_LEEWAY"`
	AuthenticationRefreshTokenCookieName             string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_NAME"`
	AuthenticationRefreshTokenCookieDomain           string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_DOMAIN"`
	AuthenticationRefreshTokenCookiePath             string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_PATH"`
	AuthenticationAccessTokenTTL                     time.Duration `mapstructure:"AUTHENTICATION_ACCESS_TOKEN_TTL"`
	AuthenticationAccessTokenSecretLength            int           `mapstructure:"AUTHENTICATION_ACCESS_TOKEN_SECRET_LENGTH"`
	AuthenticationEmailVerificationRateLimitInterval time.Duration `mapstructure:"AUTHENTICATION_EMAIL_VERIFICATION_RATE_LIMIT_INTERVAL"`
	AuthenticationEmailVerificationTokenTTL          time.Duration `mapstructure:"AUTHENTICATION_EMAIL_VERIFICATION_TOKEN_TTL"`
	AuthenticationEmailVerificationTokenSecretLength int           `mapstructure:"AUTHENTICATION_EMAIL_VERIFICATION_TOKEN_SECRET_LENGTH"`
	AuthenticationEmailVerificationURL               string        `mapstructure:"AUTHENTICATION_EMAIL_VERIFICATION_URL"`
	AuthenticationPasswordResetRateLimitInterval     time.Duration `mapstructure:"AUTHENTICATION_PASSWORD_RESET_RATE_LIMIT_INTERVAL"`
	AuthenticationPasswordResetTokenTTL              time.Duration `mapstructure:"AUTHENTICATION_PASSWORD_RESET_TOKEN_TTL"`
	AuthenticationPasswordResetTokenSecretLength     int           `mapstructure:"AUTHENTICATION_PASSWORD_RESET_TOKEN_SECRET_LENGTH"`
	AuthenticationPasswordResetURL                   string        `mapstructure:"AUTHENTICATION_PASSWORD_RESET_URL"`
	AuthenticationEmailBlocklist                     map[string]struct{}

	CaptchaEnabled            bool   `mapstructure:"CAPTCHA_ENABLED"`
	CaptchaTurnstileBaseURL   string `mapstructure:"CAPTCHA_TURNSTILE_BASE_URL"`
	CaptchaTurnstileSecretKey string `mapstructure:"CAPTCHA_TURNSTILE_SECRET_KEY"`

	TransactionalEmailsEnabled             bool   `mapstructure:"TRANSACTIONAL_EMAILS_ENABLED"`
	TransactionalEmailsDailyGlobalLimit    int    `mapstructure:"TRANSACTIONAL_EMAILS_DAILY_GLOBAL_LIMIT"`
	TransactionalEmailsSenderEmail         string `mapstructure:"TRANSACTIONAL_EMAILS_SENDER_EMAIL"`
	TransactionalEmailsSenderName          string `mapstructure:"TRANSACTIONAL_EMAILS_SENDER_NAME"`
	TransactionalEmailsScalewayAccessKeyID string `mapstructure:"TRANSACTIONAL_EMAILS_SCALEWAY_ACCESS_KEY_ID"`
	TransactionalEmailsScalewaySecretKey   string `mapstructure:"TRANSACTIONAL_EMAILS_SCALEWAY_SECRET_KEY"`
	TransactionalEmailsScalewayRegionRaw   string `mapstructure:"TRANSACTIONAL_EMAILS_SCALEWAY_REGION"`
	TransactionalEmailsScalewayRegion      scw.Region
	TransactionalEmailsScalewayProjectID   string `mapstructure:"TRANSACTIONAL_EMAILS_SCALEWAY_PROJECT_ID"`

	TasksRedisAddr     string `mapstructure:"TASKS_REDIS_ADDR"`
	TasksRedisPassword string `mapstructure:"TASKS_REDIS_PASSWORD"`
}

func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("json")

	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	// Server configuration
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "json")
	viper.SetDefault("log_time_format", "iso8601")
	viper.SetDefault("request_timeout", 60*time.Second)
	// No default for CORS Origins
	viper.SetDefault("cors_allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors_allowed_headers", []string{"Accept", "Authorization", "Content-Type", "X-Captcha-Response"})
	viper.SetDefault("cors_exposed_headers", []string{"Link"})
	viper.SetDefault("cors_allow_credentials", true)
	viper.SetDefault("cors_max_age", 5*time.Minute)
	viper.SetDefault("shutdown_timeout", 15*time.Second)
	viper.SetDefault("listen_addr", ":3333")
	viper.SetDefault("read_timeout", 0)
	viper.SetDefault("write_timeout", 0)
	viper.SetDefault("read_header_timeout", 10*time.Second)
	viper.SetDefault("idle_timeout", 0*time.Second)

	// Database configuration
	// No default for database URL
	viper.SetDefault("database_max_open_connections", 20)
	viper.SetDefault("database_max_idle_connections", 5)
	viper.SetDefault("database_max_conn_lifetime", 30*time.Minute)
	viper.SetDefault("database_max_conn_idle_time", 5*time.Minute)

	// Authentication
	viper.SetDefault("authentication_bcrypt_cost", 12)
	viper.SetDefault("authentication_timing_attack_delay", 500*time.Millisecond)
	viper.SetDefault("authentication_refresh_token_ttl", 36*time.Hour)
	viper.SetDefault("authentication_refresh_token_secret_length", 32)
	viper.SetDefault("authentication_refresh_token_leeway", 0)
	viper.SetDefault("authentication_refresh_token_cookie_name", "refresh_token")
	viper.SetDefault("authentication_refresh_token_cookie_domain", "")
	viper.SetDefault("authentication_refresh_token_cookie_path", "/account/refresh-session")
	viper.SetDefault("authentication_access_token_ttl", 5*time.Minute)
	viper.SetDefault("authentication_access_token_secret_length", 32)
	viper.SetDefault("authentication_email_verification_rate_limit_interval", 15*time.Minute)
	viper.SetDefault("authentication_email_verification_token_ttl", 15*time.Minute)
	viper.SetDefault("authentication_email_verification_token_secret_length", 32)
	viper.SetDefault("authentication_email_verification_url", "https://example.com/settings/verify-email")
	viper.SetDefault("authentication_password_reset_rate_limit_interval", 15*time.Minute)
	viper.SetDefault("authentication_password_reset_token_ttl", 15*time.Minute)
	viper.SetDefault("authentication_password_reset_token_secret_length", 32)
	viper.SetDefault("authentication_password_reset_url", "https://example.com/settings/reset-password")
	// AuthenticationEmailBlocklist is loaded from a file and parsed later

	// Captcha
	viper.SetDefault("captcha_enabled", true)
	viper.SetDefault("captcha_turnstile_base_url", "https://challenges.cloudflare.com/turnstile/v0")
	// No default for Turnstile secret key

	// Transactional Emails
	viper.SetDefault("transactional_emails_enabled", true)
	viper.SetDefault("transactional_emails_daily_global_limit", 500)
	viper.SetDefault("transactional_emails_sender_email", "noreply@example.com.com")
	viper.SetDefault("transactional_emails_sender_name", "Go API Template")
	// No default for Scaleway access key ID
	// No default for Scaleway secret key
	viper.SetDefault("transactional_emails_scaleway_region", "fr-par")
	// No default for Scaleway project ID

	// Tasks
	viper.SetDefault("tasks_redis_addr", "localhost:6379")
	viper.SetDefault("tasks_redis_password", "")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	config.TransactionalEmailsScalewayRegion = parseScalewayRegion(config.TransactionalEmailsScalewayRegionRaw)
	config.AuthenticationEmailBlocklist = loadAuthenticationEmailBlocklist()

	return config, nil
}

func parseScalewayRegion(s string) scw.Region {
	region, err := scw.ParseRegion(s)

	if err != nil {
		panic("invalid Scaleway region: " + s)
	}

	return region
}

// See https://github.com/disposable-email-domains/disposable-email-domains
func loadAuthenticationEmailBlocklist() map[string]struct{} {
	list := make(map[string]struct{})

	f, err := os.Open("./config/email-blocklist.conf")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		list[line] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return list
}
