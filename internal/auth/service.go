package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"labgrab/pkg/config"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("auth-service")

type Service struct {
	cfg    *config.AuthServiceConfig
	logger *zap.SugaredLogger
}

func (s *Service) ValidateTelegramAuthData(ctx context.Context, data *TelegramAuthData) error {
	ctx, span := tracer.Start(ctx, "auth.service.ValidateTelegramAuthData")
	defer span.End()

	_, hashSpan := tracer.Start(ctx, "auth.service.VerifyHash")
	hashErr := s.verifyHash(data)
	hashSpan.End()

	if hashErr != nil {
		hashSpan.RecordError(hashErr)
		hashSpan.SetStatus(codes.Error, hashErr.Error())
		span.RecordError(hashErr)
		span.SetStatus(codes.Error, hashErr.Error())
		s.logger.Errorw(hashErr.Error())
	}

	_, dateSpan := tracer.Start(ctx, "auth.service.VerifyAuthDate")
	dateErr := s.verifyAuthDate(data.AuthDate)
	dateSpan.End()

	if dateErr != nil {
		dateSpan.RecordError(dateErr)
		dateSpan.SetStatus(codes.Error, dateErr.Error())
		span.RecordError(hashErr)
		span.SetStatus(codes.Error, dateErr.Error())
		s.logger.Errorw(dateErr.Error())
	}

	s.logger.Info("telegram auth data verified successfully")

	return nil
}

func (s *Service) verifyHash(data *TelegramAuthData) error {
	dataCheckString := s.buildDataCheckString(data)
	key := sha256.Sum256([]byte(s.cfg.BotToken))

	h := hmac.New(sha256.New, key[:])
	h.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(hash), []byte(data.Hash)) {
		return &ErrHashIntegrity{
			ExpectedHash: data.Hash,
			ActualHash:   hash,
		}
	}
	return nil
}

func (s *Service) buildDataCheckString(data *TelegramAuthData) string {
	fields := make(map[string]string)

	fields["id"] = strconv.Itoa(data.Id)
	fields["first_name"] = data.FirstName
	fields["last_name"] = data.LastName
	fields["username"] = data.Username
	fields["photo_url"] = data.PhotoURL
	fields["auth_date"] = strconv.FormatInt(data.AuthDate.Unix(), 10)

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, fields[k]))
	}

	return strings.Join(parts, "\n")
}

func (s *Service) verifyAuthDate(authDate time.Time) error {
	currentDate := time.Now()
	if currentDate.Sub(authDate).Hours() > 24 {
		return &ErrAuthDateExpired{
			AuthDate:    authDate,
			CurrentDate: currentDate,
		}
	}
	return nil
}
