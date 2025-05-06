package config

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel             string        `mapstructure:"LOG_LEVEL"`
	LogFormat            string        `mapstructure:"LOG_FORMAT"`
	DatabaseUrl          string        `mapstructure:"DATABASE_URL"`
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
	TimingAttackDelay    time.Duration `mapstructure:"TIMING_ATTACK_DELAY"`

	AuthenticationTimingAttackDelay             time.Duration `mapstructure:"AUTHENTICATION_TIMING_ATTACK_DELAY"`
	AuthenticationRefreshTokenTTL               time.Duration `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_TTL"`
	AuthenticationRefreshTokenSecretLength      int           `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_SECRET_LENGTH"`
	AuthenticationRefreshTokenLeeway            time.Duration `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_LEEWAY"`
	AuthenticationRefreshTokenCookieName        string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_NAME"`
	AuthenticationRefreshTokenCookieDomain      string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_DOMAIN"`
	AuthenticationRefreshTokenCookiePath        string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_PATH"`
	AuthenticationRefreshTokenCookieSecure      bool          `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_SECURE"`
	AuthenticationRefreshTokenCookieHttpOnly    bool          `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_HTTP_ONLY"`
	AuthenticationRefreshTokenCookieSameSiteRaw string        `mapstructure:"AUTHENTICATION_REFRESH_TOKEN_COOKIE_SAME_SITE"`
	AuthenticationRefreshTokenCookieSameSite    http.SameSite
	AuthenticationAccessTokenTTL                time.Duration `mapstructure:"AUTHENTICATION_ACCESS_TOKEN_TTL"`
	AuthenticationAccessTokenSecretLength       int           `mapstructure:"AUTHENTICATION_ACCESS_TOKEN_SECRET_LENGTH"`

	TasksRedisAddr     string `mapstructure:"TASKS_REDIS_ADDR"`
	TasksRedisPassword string `mapstructure:"TASKS_REDIS_PASSWORD"`
}

func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("json")

	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	viper.SetDefault("log_level", "debug")
	viper.SetDefault("log_format", "json")
	viper.SetDefault("request_timeout", 60*time.Second)
	viper.SetDefault("cors_allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("cors_allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors_allowed_headers", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	viper.SetDefault("cors_exposed_headers", []string{"Link"})
	viper.SetDefault("cors_allow_credentials", false)
	viper.SetDefault("cors_max_age", 5*time.Minute)
	viper.SetDefault("shutdown_timeout", 15*time.Second)
	viper.SetDefault("listen_addr", ":3333")
	viper.SetDefault("read_timeout", 0)
	viper.SetDefault("write_timeout", 0)
	viper.SetDefault("read_header_timeout", 10*time.Second)
	viper.SetDefault("idle_timeout", 0*time.Second)
	viper.SetDefault("timing_attack_delay", 500*time.Millisecond)

	viper.SetDefault("authentication_timing_attack_delay", 500*time.Millisecond)
	viper.SetDefault("authentication_refresh_token_ttl", 36*time.Hour)
	viper.SetDefault("authentication_refresh_token_secret_length", 32)
	viper.SetDefault("authentication_refresh_token_leeway", 0)
	viper.SetDefault("authentication_refresh_token_cookie_name", "refresh_token")
	viper.SetDefault("authentication_refresh_token_cookie_domain", "")
	viper.SetDefault("authentication_refresh_token_cookie_path", "/sessions/refresh")
	viper.SetDefault("authentication_refresh_token_cookie_secure", true)
	viper.SetDefault("authentication_refresh_token_cookie_http_only", true)
	viper.SetDefault("authentication_refresh_token_cookie_same_site", "strict")
	viper.SetDefault("authentication_access_token_ttl", 15*time.Minute)
	viper.SetDefault("authentication_access_token_secret_length", 32)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
