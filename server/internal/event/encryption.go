package event

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/errors"

	"github.com/google/uuid"
)

func (s *Service) DecryptUserData(data *DBUserData) (*DecryptedUserData, error) {
	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		return nil, &errors.ErrServiceProcedure{
			Procedure: "DecryptUserData",
			Step:      "DecryptDEK",
			Err:       err,
		}
	}

	password, err := decryptWithDEK(data.DikidiPassword, rawDEK)
	if err != nil {
		return nil, &errors.ErrServiceProcedure{
			Procedure: "DecryptUserData",
			Step:      "DecryptPassword",
			Err:       err,
		}
	}

	session, err := decryptPtrWithDEK(data.Session, rawDEK)
	if err != nil {
		return nil, &errors.ErrServiceProcedure{
			Procedure: "DecryptUserData",
			Step:      "DecryptSession",
			Err:       err,
		}
	}

	token, err := decryptPtrWithDEK(data.Token, rawDEK)
	if err != nil {
		return nil, &errors.ErrServiceProcedure{
			Procedure: "DecryptUserData",
			Step:      "DecryptToken",
			Err:       err,
		}
	}

	cookies, err := decryptPtrWithDEK(data.Cookies, rawDEK)
	if err != nil {
		return nil, &errors.ErrServiceProcedure{
			Procedure: "DecryptUserData",
			Step:      "DecryptCookies",
			Err:       err,
		}
	}

	return &DecryptedUserData{
		UserUUID:          data.UserUUID,
		DikidiPhoneNumber: data.DikidiPhoneNumber,
		DikidiPassword:    password,
		Session:           session,
		Token:             token,
		Cookies:           cookies,
	}, nil
}

func (s *Service) EncryptPassword(password string, userUUID uuid.UUID) (string, string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", "", err
	}

	encPass, err := encryptWithDEK(password, key)
	if err != nil {
		return "", "", err
	}

	eDEK := s.kekGCM.Seal(nil, nil, key, []byte(userUUID.String()))

	return encPass, base64.StdEncoding.EncodeToString(eDEK), nil
}

func (s *Service) DecryptDEK(encDEK string, userUUID uuid.UUID) ([]byte, error) {
	eDEK, err := base64.StdEncoding.DecodeString(encDEK)
	if err != nil {
		return nil, err
	}
	return s.kekGCM.Open(nil, nil, eDEK, []byte(userUUID.String()))
}

func (s *Service) EncryptAndSaveCookies(ctx context.Context, userUUID uuid.UUID, rawDEK []byte, cookies *dikidi.ClientCookies, session string) error {
	encSession, err := encryptWithDEK(session, rawDEK)
	if err != nil {
		return &errors.ErrServiceProcedure{Procedure: "encryptAndSaveCookies", Step: "EncryptSession", Err: err}
	}

	encToken, err := encryptWithDEK(*cookies.Token, rawDEK)
	if err != nil {
		return &errors.ErrServiceProcedure{Procedure: "encryptAndSaveCookies", Step: "EncryptToken", Err: err}
	}

	encCookies, err := encryptWithDEK(cookies.All, rawDEK)
	if err != nil {
		return &errors.ErrServiceProcedure{Procedure: "encryptAndSaveCookies", Step: "EncryptCookies", Err: err}
	}

	return s.repo.SetUserCookies(ctx, userUUID, &DBUserCookies{
		Session: &encSession,
		Token:   &encToken,
		Cookies: &encCookies,
	})
}

func encryptWithDEK(plaintext string, rawDEK []byte) (string, error) {
	dekCipher, err := aes.NewCipher(rawDEK)
	if err != nil {
		return "", err
	}
	dekGCM, err := cipher.NewGCMWithRandomNonce(dekCipher)
	if err != nil {
		return "", err
	}
	encrypted := dekGCM.Seal(nil, nil, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func decryptWithDEK(ciphertext string, rawDEK []byte) (string, error) {
	eCipher, err := base64.StdEncoding.DecodeString(ciphertext)
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
	raw, err := dekGCM.Open(nil, nil, eCipher, nil)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decryptPtrWithDEK(ciphertext *string, rawDEK []byte) (*string, error) {
	if ciphertext == nil {
		return nil, nil
	}
	plain, err := decryptWithDEK(*ciphertext, rawDEK)
	if err != nil {
		return nil, err
	}
	return &plain, nil
}
