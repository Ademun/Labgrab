package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/mask"
	"labgrab/pkg/config"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("auth-service")

type Service struct {
	repo   *Repo
	cache  *redis.Client
	client *dikidi.Client
	cfg    *config.AuthServiceConfig
	kekGCM cipher.AEAD
}

func NewService(
	repo *Repo,
	cache *redis.Client,
	client *dikidi.Client,
	cfg *config.AuthServiceConfig,
	encCfg *config.EncryptionConfig,
) (*Service, error) {
	key, err := hex.DecodeString(encCfg.PasswordKEK)
	if err != nil {
		return nil, fmt.Errorf("auth service: new service: decode kek key: %w", err)
	}

	kekCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("auth service: new service: create aes cipher: %w", err)
	}

	kekGCM, err := cipher.NewGCMWithRandomNonce(kekCipher)
	if err != nil {
		return nil, fmt.Errorf("auth service: new service: create gcm: %w", err)
	}

	return &Service{
		repo:   repo,
		cache:  cache,
		client: client,
		cfg:    cfg,
		kekGCM: kekGCM,
	}, nil
}

// --- Session management ---

func (s *Service) CreateSession(ctx context.Context, userUUID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth service: create session: generate random bytes: %w", err)
	}

	session := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	key := fmt.Sprintf("%s:%s", s.cfg.KeyPrefix, session)

	if err := s.cache.Set(ctx, key, userUUID.String(), time.Hour*24).Err(); err != nil {
		return "", fmt.Errorf("auth service: create session: store session key: %w", err)
	}

	return session, nil
}

func (s *Service) ValidateSession(ctx context.Context, session string) error {
	key := fmt.Sprintf("%s:%s", s.cfg.KeyPrefix, session)

	n, err := s.cache.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("auth service: validate session: check session key: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("auth service: validate session: session not found")
	}

	return nil
}

func (s *Service) GetSessionData(ctx context.Context, session string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", s.cfg.KeyPrefix, session)

	result, err := s.cache.Get(ctx, key).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth service: get session data: retrieve session: %w", err)
	}

	userUUID, err := uuid.Parse(result)
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth service: get session data: parse uuid: %w", err)
	}

	return userUUID, nil
}

// --- Telegram auth ---

func (s *Service) ValidateTelegramAuthData(ctx context.Context, data *TelegramAuthData) error {
	ctx, span := tracer.Start(ctx, "auth_service.validate_telegram_auth_data")
	defer span.End()

	if err := s.verifyHash(data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to verify hash")
		return fmt.Errorf("auth service: validate telegram auth data: verify hash: %w", err)
	}

	if err := s.verifyAuthDate(data.AuthDate); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to verify auth date")
		return fmt.Errorf("auth service: validate telegram auth data: verify auth date: %w", err)
	}

	span.SetStatus(codes.Ok, "")
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
	fields := map[string]string{
		"id":        strconv.Itoa(data.Id),
		"auth_date": strconv.Itoa(data.AuthDate),
	}

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
	now := time.Now()
	authDateTime := time.Unix(int64(authDate), 0)
	if now.Sub(authDateTime).Hours() > 24 {
		return &ErrAuthDateExpired{
			AuthDate:    authDateTime,
			CurrentDate: now,
		}
	}
	return nil
}

// --- Dikidi credential management ---

func (s *Service) CreateUserData(ctx context.Context, req *CreateUserDataReq) error {
	ctx, span := tracer.Start(ctx, "auth_service.create_user_data")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	pass, dek, err := s.EncryptPassword(req.DikidiPassword, req.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt password")
		return fmt.Errorf("auth service: create user data: encrypt password: %w", err)
	}

	if err := s.repo.CreateUserData(ctx, &DBUserData{
		UserUUID:          req.UserUUID,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
		DikidiPassword:    pass,
		DEK:               dek,
	}, req.Tx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user data")
		return fmt.Errorf("auth service: create user data: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) AuthUser(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "auth_service.auth_user")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	data, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return fmt.Errorf("auth service: auth user: get user data: %w", err)
	}

	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt dek")
		return fmt.Errorf("auth service: auth user: decrypt dek: %w", err)
	}

	password, err := decryptWithDEK(data.DikidiPassword, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt password")
		return fmt.Errorf("auth service: auth user: decrypt password: %w", err)
	}

	client := mask.CreateRandomHTTPClient()

	telegramCSRF, err := s.client.AcquireTelegramCSRFToken(ctx, client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire telegram csrf token")
		return fmt.Errorf("auth service: auth user: acquire telegram csrf token: %w", err)
	}

	mask.Jitter(2000, 3000)

	csrf, err := s.client.AcquireCSRFToken(ctx, client, dikidi.CSRFTokenRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		TelegramCSRFToken: telegramCSRF,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire csrf token")
		return fmt.Errorf("auth service: auth user: acquire csrf token: %w", err)
	}

	mask.Jitter(5000, 8000)

	if err = s.client.SendAuthRequest(ctx, client, dikidi.AuthRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		Password:          password,
		CSRFToken:         csrf,
		TelegramCSRFToken: telegramCSRF,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to send auth request")
		return fmt.Errorf("auth service: auth user: send auth request: %w", err)
	}

	cookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return fmt.Errorf("auth service: auth user: acquire client cookies: %w", err)
	}

	if cookies.CookieName == nil || cookies.Token == nil {
		err = errors.New("cookie_name or token is missing")
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return fmt.Errorf("auth service: auth user: acquire client cookies: %w", err)
	}

	session, err := s.client.AcquireSessionID(*cookies.CookieName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session id")
		return fmt.Errorf("auth service: auth user: acquire session id: %w", err)
	}

	if err = s.EncryptAndSaveCookies(ctx, data.UserUUID, rawDEK, cookies, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt and save cookies")
		return fmt.Errorf("auth service: auth user: encrypt and save cookies: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) RefreshUserCookies(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "auth_service.refresh_user_cookies")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	eData, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return fmt.Errorf("auth service: refresh user cookies: get user data: %w", err)
	}

	rawDEK, err := s.DecryptDEK(eData.DEK, eData.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt dek")
		return fmt.Errorf("auth service: refresh user cookies: decrypt dek: %w", err)
	}

	existingCookies, err := decryptPtrWithDEK(eData.Cookies, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt existing cookies")
		return fmt.Errorf("auth service: refresh user cookies: decrypt cookies: %w", err)
	}

	client, err := mask.CreateClientWithCookies(existingCookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create http client with cookies")
		return fmt.Errorf("auth service: refresh user cookies: create http client with cookies: %w", err)
	}

	if err = s.client.RenewCookies(ctx, client); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to renew cookies")
		return fmt.Errorf("auth service: refresh user cookies: renew cookies: %w", err)
	}

	newCookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return fmt.Errorf("auth service: refresh user cookies: acquire client cookies: %w", err)
	}

	if newCookies.CookieName == nil || newCookies.Token == nil {
		err = errors.New("cookie_name or token is missing")
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return fmt.Errorf("auth service: refresh user cookies: acquire client cookies: %w", err)
	}

	session, err := s.client.AcquireSessionID(*newCookies.CookieName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session id")
		return fmt.Errorf("auth service: refresh user cookies: acquire session id: %w", err)
	}

	if err = s.EncryptAndSaveCookies(ctx, eData.UserUUID, rawDEK, newCookies, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt and save cookies")
		return fmt.Errorf("auth service: refresh user cookies: encrypt and save cookies: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) GetUserCredentials(ctx context.Context, userUUID uuid.UUID) (*GetUserCredentialsRes, error) {
	ctx, span := tracer.Start(ctx, "auth_service.get_user_credentials")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	data, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return nil, fmt.Errorf("auth service: get user credentials: get user data: %w", err)
	}

	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt dek")
		return nil, fmt.Errorf("auth service: get user credentials: decrypt dek: %w", err)
	}

	session, err := decryptPtrWithDEK(data.Session, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt session")
		return nil, fmt.Errorf("auth service: get user credentials: decrypt session: %w", err)
	}

	token, err := decryptPtrWithDEK(data.Token, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt token")
		return nil, fmt.Errorf("auth service: get user credentials: decrypt token: %w", err)
	}

	cookies, err := decryptPtrWithDEK(data.Cookies, rawDEK)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt cookies")
		return nil, fmt.Errorf("auth service: get user credentials: decrypt cookies: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return &GetUserCredentialsRes{
		DikidiPhoneNumber: data.DikidiPhoneNumber,
		Session:           session,
		Token:             token,
		Cookies:           cookies,
	}, nil
}

func sanitizePhoneNumber(phoneNumber string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phoneNumber)
}
