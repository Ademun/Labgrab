package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	InfraConfig               InfraConfig
	APIClientConfig           DikidiClientConfig        `yaml:"dikidi_client"`
	PollingServiceConfig      PollingServiceConfig      `yaml:"polling_service"`
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
