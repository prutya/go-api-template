package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel       string        `mapstructure:"LOG_LEVEL"`
	LogTimeFormat  string        `mapstructure:"LOG_TIME_FORMAT"`
	RequestTimeout time.Duration `mapstructure:"REQUEST_TIMEOUT"`
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
