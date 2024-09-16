package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel             string        `mapstructure:"LOG_LEVEL"`
	LogTimeFormat        string        `mapstructure:"LOG_TIME_FORMAT"`
	RequestTimeout       time.Duration `mapstructure:"REQUEST_TIMEOUT"`
	CorsAllowedOrigins   []string      `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CorsAllowedMethods   []string      `mapstructure:"CORS_ALLOWED_METHODS"`
	CorsAllowedHeaders   []string      `mapstructure:"CORS_ALLOWED_HEADERS"`
	CorsExposedHeaders   []string      `mapstructure:"CORS_EXPOSED_HEADERS"`
	CorsAllowCredentials bool          `mapstructure:"CORS_ALLOW_CREDENTIALS"`
	CorsMaxAge           time.Duration `mapstructure:"CORS_MAX_AGE"`
}

func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("json")

	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
