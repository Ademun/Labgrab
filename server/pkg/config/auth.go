package config

type AuthServiceConfig struct {
	BotToken    string
	KeyPrefix   string `yaml:"key_prefix"`
	PasswordKEK string `envconfig:"PASSWORD_KEK"`
}
