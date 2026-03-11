package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InfraConfig               InfraConfig
	TelegramConfig            TelegramConfig
	APIClientConfig           DikidiClientConfig        `yaml:"dikidi_client"`
	AuthServiceConfig         AuthServiceConfig         `yaml:"auth_service"`
	PollingServiceConfig      PollingServiceConfig      `yaml:"polling_service"`
	SubscriptionServiceConfig SubscriptionServiceConfig `yaml:"subscription_service"`
}

func Load() (*Config, error) {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	cfg.AuthServiceConfig.BotToken = cfg.TelegramConfig.BotToken

	return &cfg, nil
}
