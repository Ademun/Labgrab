package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	InfraConfig               InfraConfig
	APIClientConfig           APIClientConfig           `yaml:"api_client"`
	SubscriptionServiceConfig SubscriptionServiceConfig `yaml:"subscription_service"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file, %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config, %v", err)
	}

	return &config, nil
}
