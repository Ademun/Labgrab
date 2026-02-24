package lab_enrollment

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"labgrab/internal/shared/errors"
	"labgrab/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("lab_enrollment-service")

type Service struct {
	repo   *Repo
	cfg    *config.EncryptionConfig
	kekGCM cipher.AEAD
}

func NewService(repo *Repo, cfg *config.EncryptionConfig) (*Service, error) {
	kekCipher, err := aes.NewCipher([]byte(cfg.PasswordKEK))
	if err != nil {
		return nil, err
	}

	kekGCM, err := cipher.NewGCMWithRandomNonce(kekCipher)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo:   repo,
		cfg:    cfg,
		kekGCM: kekGCM,
	}, nil
}

func (s *Service) CreateUserData(ctx context.Context, req *CreateuserDataReq) error {
	ctx, span := tracer.Start(ctx, "lab_enrollment_service.create_user_data")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
	)

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "Key generation",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fill AES key with random bytes")
		return err
	}

	dekCipher, err := aes.NewCipher(key)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "Cipher initialisation",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to initialise AES DEK cipher")
		return err
	}

	dekGCM, err := cipher.NewGCMWithRandomNonce(dekCipher)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUserData",
			Step:      "GCM wrapping",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to wrap DEK cipher with GCM")
		return err
	}

	ePass := dekGCM.Seal(nil, nil, []byte(req.DikidiPassword), nil)

	eDEK := s.kekGCM.Seal(nil, nil, key, []byte(req.UserUUID.String()))

	dbUserData := &DBUserData{
		UUID:              req.UserUUID,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
		DikidiPassword:    base64.StdEncoding.EncodeToString(ePass),
		PasswordDEK:       base64.StdEncoding.EncodeToString(eDEK),
	}

	if err := s.repo.CreateUserData(ctx, dbUserData, req.tx); err != nil {
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
