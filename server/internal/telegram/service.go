package telegram

import (
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type Service struct {
	bot *bot.Bot

	logger *zap.SugaredLogger
}
