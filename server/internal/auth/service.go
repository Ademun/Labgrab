package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"labgrab/internal/shared/errors"
	"labgrab/pkg/config"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("auth-service")

type Service struct {
	cache *redis.Client
	cfg   *config.AuthServiceConfig
}

func NewService(cache *redis.Client, cfg *config.AuthServiceConfig) *Service {
	return &Service{
		cache: cache,
		cfg:   cfg,
	}
}

func (s *Service) CreateSession(ctx context.Context, userUUID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", &errors.ErrServiceProcedure{
			Procedure: "CreateSession",
			Step:      "Random number generation",
			Err:       err,
		}
	}

	session := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	key := fmt.Sprintf("%s:%s", s.cfg.KeyPrefix, session)

	if err := s.cache.Set(ctx, key, userUUID, time.Hour*24).Err(); err != nil {
		return "", &errors.ErrServiceProcedure{
			Procedure: "CreateSession",
			Step:      "Store session key",
			Err:       err,
		}
	}

	return session, nil
}

func (s *Service) ValidateSession(ctx context.Context, session string) error {
	key := fmt.Sprintf("%s:%s", s.cfg.KeyPrefix, session)
	if err := s.cache.Exists(ctx, key).Err(); err != nil {
		return &errors.ErrServiceProcedure{
			Procedure: "ValidateSession",
			Step:      "Check session key",
			Err:       err,
		}
	}
	return nil
}

func (s *Service) GetSessionData(ctx context.Context, session string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", s.cfg.KeyPrefix, session)

	result, err := s.cache.Get(ctx, key).Bytes()
	if err != nil {
		return uuid.Nil, &errors.ErrServiceProcedure{
			Procedure: "GetSessionData",
			Step:      "Retrieve session data",
			Err:       err,
		}
	}

	userUUID, err := uuid.FromBytes(result)
	if err != nil {
		return uuid.Nil, &errors.ErrServiceProcedure{
			Procedure: "GetSessionData",
			Step:      "Unmarshal session data",
			Err:       err,
		}
	}

	return userUUID, nil
}

func (s *Service) ValidateTelegramAuthData(ctx context.Context, data *TelegramAuthData) error {
	ctx, span := tracer.Start(ctx, "auth.service.ValidateTelegramAuthData")
	defer span.End()

	err := s.verifyHash(data)

	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "ValidateTelegramAuthData",
			Step:      "Hash verification",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	err = s.verifyAuthDate(data.AuthDate)

	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "ValidateTelegramAuthData",
			Step:      "Auth date verification",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	fields["auth_date"] = strconv.Itoa(data.AuthDate)

	if data.FirstName != "" {
		fields["first_name"] = data.FirstName
	}
	if data.LastName != "" {
		fields["last_name"] = data.LastName
	}
	if data.Username != "" {
		fields["username"] = data.Username
	}
	if data.PhotoURL != "" {
		fields["photo_url"] = data.PhotoURL
	}

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

func (s *Service) verifyAuthDate(authDate int) error {
	currentDate := time.Now()
	authDateTime := time.Unix(int64(authDate), 0)
	if currentDate.Sub(authDateTime).Hours() > 24 {
		return &ErrAuthDateExpired{
			AuthDate:    authDateTime,
			CurrentDate: currentDate,
		}
	}
	return nil
}
