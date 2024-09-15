package config

import "github.com/spf13/viper"

type Config struct {
	LogLevel      string `mapstructure:"LOG_LEVEL"`
	LogTimeFormat string `mapstructure:"LOG_TIME_FORMAT"`
}

func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app.config")
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
