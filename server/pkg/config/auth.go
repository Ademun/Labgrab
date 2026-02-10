package config

type AuthServiceConfig struct {
	BotToken  string `envconfig:"BOT_TOKEN"`
	KeyPrefix string `yaml:"key_prefix"`
}
