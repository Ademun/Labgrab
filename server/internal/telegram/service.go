package telegram

import (
	"context"
	"fmt"
	"labgrab/internal/shared/domain"
	"labgrab/pkg/config"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Service struct {
	bot *bot.Bot
	cfg *config.TelegramConfig
}

func NewService(cfg *config.TelegramConfig) (*Service, error) {
	b, err := bot.New(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	return &Service{
		bot: b,
		cfg: cfg,
	}, nil
}

func (s *Service) Start(ctx context.Context) {
	s.bot.Start(ctx)
}

func (s *Service) NotifyEvent(ctx context.Context, req NotifyEventReq) error {
	msg := createNotificationMessage(req)
	_, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      req.UserID,
		Text:        msg,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: createLinkButton(req.PageURL),
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
	})

	return err
}

func (s *Service) NotifyEnrollment(ctx context.Context, req NotifyEnrollmentReq) error {
	msg := createEnrollmentMessage(req)
	_, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    req.UserID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})

	return err
}

func createNotificationMessage(req NotifyEventReq) string {
	var sb strings.Builder
	sb.WriteString(bold("🔥 Появилась запись!"))
	sb.WriteString(breakLine(2))
	var spotString string
	if req.Spot != nil {
		spotString = fmt.Sprintf("(%d-е место)", *req.Spot)
	}
	sb.WriteString(bold(fmt.Sprintf("📚 Работа №%d. %s. %s %s", req.LabNumber, req.LabType.RU(), req.LabName, spotString)))
	sb.WriteString(breakLine(2))
	sb.WriteString(bold(fmt.Sprintf("%s %s", req.LabTopic.Icon(), req.LabTopic.RU())))
	sb.WriteString(breakLine(2))
	sb.WriteString(bold(fmt.Sprintf("🚪 Аудитория №%d", req.LabAuditorium)))
	sb.WriteString(breakLine(2))
	sb.WriteString(bold("📅 Расписание:"))
	sb.WriteString(breakLine(1))
	for dateTime, info := range req.Schedule {
		sb.WriteString(bold(ident(formatDateTime(dateTime), 1)))
		sb.WriteString(breakLine(1))
		for lesson, teachers := range info {
			str := fmt.Sprintf("%s: %s", formatLesson(lesson), strings.Join(teachers, ", "))
			sb.WriteString(bold(ident(str, 2)))
			sb.WriteString(breakLine(1))
		}
		sb.WriteString(breakLine(1))
	}
	return sb.String()
}

func createEnrollmentMessage(req NotifyEnrollmentReq) string {
	var sb strings.Builder
	sb.WriteString(bold("⭐ Успешная запись!"))
	sb.WriteString(breakLine(2))
	var spotString string
	if req.Spot != nil {
		spotString = fmt.Sprintf("(%d-е место)", *req.Spot)
	}
	sb.WriteString(bold(fmt.Sprintf("📚 Работа №%d. %s. %s %s", req.LabNumber, req.LabType.RU(), req.LabName, spotString)))
	sb.WriteString(breakLine(2))
	sb.WriteString(bold(fmt.Sprintf("%s %s", req.LabTopic.Icon(), req.LabTopic.RU())))
	sb.WriteString(breakLine(2))
	sb.WriteString(bold(fmt.Sprintf("🚪 Аудитория №%d", req.LabAuditorium)))
	sb.WriteString(breakLine(2))
	sb.WriteString(fmt.Sprintf("📅 Дата работы: %s %s", formatDateTime(req.Date), formatLesson(req.Lesson)))
	return sb.String()
}

func createLinkButton(link string) models.ReplyMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "🔗 ЗАПИСАТЬСЯ", URL: link}},
		},
	}
}

func breakLine(n int) string {
	return strings.Repeat("\n", n)
}

func bold(s string) string {
	return fmt.Sprintf("<b>%s</b>", s)
}

func ident(s string, n int) string {
	return fmt.Sprintf("%s%s", strings.Repeat("\t", n), s)
}

func formatDateTime(dateTime time.Time) string {
	delta := dateTime.Sub(time.Now().In(dateTime.Location())).Hours()
	switch {
	case delta < 24:
		return "Сегодня"
	case (24 <= delta) && (delta < 48):
		return "Завтра"
	default:
		day := dateTime.Day()
		ruMonth := ruMonths[dateTime.Month()]
		ruWeekday := ruWeekdays[dateTime.Weekday()]
		return fmt.Sprintf("%d %s (%s)", day, ruMonth, ruWeekday)
	}
}

func formatLesson(lesson domain.Lesson) string {
	icon, found := lessonIcons[lesson]
	if !found {
		icon = strconv.Itoa(int(lesson))
	}
	return fmt.Sprintf("%s пара", icon)
}
