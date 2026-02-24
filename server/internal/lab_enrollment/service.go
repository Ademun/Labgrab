package lab_enrollment

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/errors"
	"labgrab/internal/shared/mask"
	"labgrab/pkg/config"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("lab_enrollment-service")

type Service struct {
	repo   *Repo
	client *dikidi.Client
	cfg    *config.EncryptionConfig
	kekGCM cipher.AEAD
}

func NewService(repo *Repo, client *dikidi.Client, cfg *config.EncryptionConfig) (*Service, error) {
	key, err := hex.DecodeString(cfg.PasswordKEK)
	if err != nil {
		return nil, err
	}

	kekCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	kekGCM, err := cipher.NewGCMWithRandomNonce(kekCipher)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo:   repo,
		client: client,
		cfg:    cfg,
		kekGCM: kekGCM,
	}, nil
}

func (s *Service) CreateUserData(ctx context.Context, req *CreateUserDataReq) error {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.create_user_data")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
	)

	pass, dek, err := s.EncryptPassword(req.DikidiPassword, req.UserUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "EncryptPassword",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encrypt password")
		return err
	}

	dbUserData := &DBUserData{
		UserUUID:          req.UserUUID,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
		DikidiPassword:    pass,
		PasswordDEK:       dek,
	}

	if err := s.repo.CreateUserData(ctx, dbUserData, req.Tx); err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user data in repository")
		return err
	}

	return nil
}

func (s *Service) AuthUser(ctx context.Context, userUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.auth_user")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
	)

	data, err := s.repo.GetUserData(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user data")
		return err
	}

	password, err := s.DecryptPassword(data.DikidiPassword, data.PasswordDEK, data.UserUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "DecryptPassword",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decrypt password")
		return err
	}

	client := s.CreateRandomHTTPClient()

	telegramCSRF, err := s.client.AcquireTelegramCSRFToken(ctx, client)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "AcquireTelegramCSRFToken",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire telegram csrf")
		return err
	}

	mask.Jitter(2000, 3000)
	csrf, err := s.client.AcquireCSRFToken(ctx, client, dikidi.CSRFTokenRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		TelegramCSRFToken: telegramCSRF,
	})
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "AcquireCSRFToken",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire CSRF token")
		return err
	}

	mask.Jitter(5000, 8000)
	err = s.client.SendAuthRequest(ctx, client, dikidi.AuthRequest{
		PhoneNumber:       sanitizePhoneNumber(data.DikidiPhoneNumber),
		Password:          password,
		CSRFToken:         csrf,
		TelegramCSRFToken: telegramCSRF,
	})
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "SendAuthRequest",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to send auth request")
		return err
	}

	cookies, err := s.client.AcquireClientCookies(client)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "AcquireClientCookies",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire client cookies")
		return err
	}

	if cookies.CookieName == nil || cookies.Token == nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "AcquireClientCookies",
			Err:       fmt.Errorf("no cookie_name or token found"),
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "bad client cookies")
		return err
	}

	session, err := s.client.AcquireSessionID(*cookies.CookieName)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "AcquireSessionID",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to acquire session ID")
		return err
	}

	dbCookies := &DBUserCookies{
		Session:      &session,
		Token:        cookies.Token,
		NoiseCookies: &cookies.All,
	}
	if err := s.repo.SetUserCookies(ctx, data.UserUUID, dbCookies); err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "AuthUser",
			Step:      "SetUserCookies",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to set user cookies")
		return err
	}

	return nil
}

func (s *Service) EncryptPassword(password string, userUUID uuid.UUID) (string, string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", "", err
	}

	dekCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	dekGCM, err := cipher.NewGCMWithRandomNonce(dekCipher)
	if err != nil {
		return "", "", err
	}

	ePass := dekGCM.Seal(nil, nil, []byte(password), nil)

	eDEK := s.kekGCM.Seal(nil, nil, key, []byte(userUUID.String()))

	return base64.StdEncoding.EncodeToString(ePass), base64.StdEncoding.EncodeToString(eDEK), nil
}

func (s *Service) DecryptPassword(password string, dek string, userUUID uuid.UUID) (string, error) {
	ePass, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", err
	}

	eDEK, err := base64.StdEncoding.DecodeString(dek)
	if err != nil {
		return "", err
	}

	rawDEK, err := s.kekGCM.Open(nil, nil, eDEK, []byte(userUUID.String()))
	if err != nil {
		return "", err
	}

	dekCipher, err := aes.NewCipher(rawDEK)
	if err != nil {
		return "", err
	}

	dekGCM, err := cipher.NewGCMWithRandomNonce(dekCipher)
	if err != nil {
		return "", err
	}

	rawPass, err := dekGCM.Open(nil, nil, ePass, nil)
	if err != nil {
		return "", err
	}

	return string(rawPass), nil
}

func (s *Service) CreateRandomHTTPClient() *req.Client {
	client := req.NewClient().
		ImpersonateChrome().
		EnableAutoDecode().
		SetCommonHeaders(map[string]string{
			"User-Agent":                "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Mobile Safari/537.36",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
			"Accept-Encoding":           "gzip, deflate, br, zstd",
			"Sec-Ch-Ua":                 `"Chromium";v="145", "Google Chrome";v="145", "Not/A)Brand";v="99"`,
			"Sec-Ch-Ua-Mobile":          "?1",
			"Sec-Ch-Ua-Platform":        `"Android"`,
			"Upgrade-Insecure-Requests": "1",
		})

	return client
}

func sanitizePhoneNumber(phoneNumber string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phoneNumber)
}
