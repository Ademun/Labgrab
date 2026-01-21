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
)

type Service struct {
	cfg *config.AuthServiceConfig
}

func (s *Service) ValidateTelegramAuthData(ctx context.Context, data *TelegramAuthData) error {

}

func (s *Service) verifyHash(data *TelegramAuthData) bool {
	dataCheckString := s.buildDataCheckString(data)
	key := sha256.Sum256([]byte(s.cfg.BotToken))

	h := hmac.New(sha256.New, key[:])
	h.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(hash), []byte(data.Hash))
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
