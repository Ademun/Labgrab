package config

type TelegramConfig struct {
	BotToken string `envconfig:"BOT_TOKEN"`
}
