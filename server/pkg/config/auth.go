package config

type AuthServiceConfig struct {
	BotToken  string `env:"BOT_TOKEN"`
	KeyPrefix string `yaml:"key_prefix"`
}
