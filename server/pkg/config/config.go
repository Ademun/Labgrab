package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InfraConfig               InfraConfig
	APIClientConfig           DikidiClientConfig `yaml:"dikidi_client"`
	AuthServiceConfig         AuthServiceConfig
	PollingServiceConfig      PollingServiceConfig      `yaml:"polling_service"`
	SubscriptionServiceConfig SubscriptionServiceConfig `yaml:"subscription_service"`
}

func Load() (*Config, error) {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	err = envconfig.Process("", &config)

	return &config, nil
}
