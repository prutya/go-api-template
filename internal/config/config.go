package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel             string        `mapstructure:"LOG_LEVEL"`
	LogTimeFormat        string        `mapstructure:"LOG_TIME_FORMAT"`
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
}

func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("json")

	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	viper.SetDefault("log_level", "debug")
	viper.SetDefault("log_time_format", "iso8601")
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

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
