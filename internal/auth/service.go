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
)

type Service struct {
	cfg *config.AuthServiceConfig
}

func (s *Service) ValidateTelegramAuthData(ctx context.Context, data *TelegramAuthData) error {
	if err := s.verifyHash(data); err != nil {
		return err
	}

	if err := s.verifyAuthDate(data.AuthDate); err != nil {
		return err
	}

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
