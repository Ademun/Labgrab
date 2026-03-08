package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"labgrab/internal/shared/api/dikidi"

	"github.com/google/uuid"
)

func (s *Service) DecryptUserData(data *DBUserData) (*DecryptedUserData, error) {
	rawDEK, err := s.DecryptDEK(data.DEK, data.UserUUID)
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt user data: decrypt dek: %w", err)
	}

	password, err := decryptWithDEK(data.DikidiPassword, rawDEK)
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt user data: decrypt password: %w", err)
	}

	session, err := decryptPtrWithDEK(data.Session, rawDEK)
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt user data: decrypt session: %w", err)
	}

	token, err := decryptPtrWithDEK(data.Token, rawDEK)
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt user data: decrypt token: %w", err)
	}

	cookies, err := decryptPtrWithDEK(data.Cookies, rawDEK)
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt user data: decrypt cookies: %w", err)
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
		return "", "", fmt.Errorf("auth service: encrypt password: generate dek: %w", err)
	}

	encPass, err := encryptWithDEK(password, key)
	if err != nil {
		return "", "", fmt.Errorf("auth service: encrypt password: encrypt with dek: %w", err)
	}

	eDEK := s.kekGCM.Seal(nil, nil, key, []byte(userUUID.String()))

	return encPass, base64.StdEncoding.EncodeToString(eDEK), nil
}

func (s *Service) DecryptDEK(encDEK string, userUUID uuid.UUID) ([]byte, error) {
	eDEK, err := base64.StdEncoding.DecodeString(encDEK)
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt dek: base64 decode: %w", err)
	}

	raw, err := s.kekGCM.Open(nil, nil, eDEK, []byte(userUUID.String()))
	if err != nil {
		return nil, fmt.Errorf("auth service: decrypt dek: gcm open: %w", err)
	}

	return raw, nil
}

func (s *Service) EncryptAndSaveCookies(ctx context.Context, userUUID uuid.UUID, rawDEK []byte, cookies *dikidi.ClientCookies, session string) error {
	encSession, err := encryptWithDEK(session, rawDEK)
	if err != nil {
		return fmt.Errorf("auth service: encrypt and save cookies: encrypt session: %w", err)
	}

	encToken, err := encryptWithDEK(*cookies.Token, rawDEK)
	if err != nil {
		return fmt.Errorf("auth service: encrypt and save cookies: encrypt token: %w", err)
	}

	encCookies, err := encryptWithDEK(cookies.All, rawDEK)
	if err != nil {
		return fmt.Errorf("auth service: encrypt and save cookies: encrypt cookies: %w", err)
	}

	if err = s.repo.SetUserCookies(ctx, userUUID, &DBUserCookies{
		Session: &encSession,
		Token:   &encToken,
		Cookies: &encCookies,
	}); err != nil {
		return fmt.Errorf("auth service: encrypt and save cookies: repository call: %w", err)
	}

	return nil
}

func encryptWithDEK(plaintext string, rawDEK []byte) (string, error) {
	dekCipher, err := aes.NewCipher(rawDEK)
	if err != nil {
		return "", fmt.Errorf("encrypt with dek: create cipher: %w", err)
	}

	dekGCM, err := cipher.NewGCMWithRandomNonce(dekCipher)
	if err != nil {
		return "", fmt.Errorf("encrypt with dek: create gcm: %w", err)
	}

	encrypted := dekGCM.Seal(nil, nil, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func decryptWithDEK(ciphertext string, rawDEK []byte) (string, error) {
	eCipher, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decrypt with dek: base64 decode: %w", err)
	}

	dekCipher, err := aes.NewCipher(rawDEK)
	if err != nil {
		return "", fmt.Errorf("decrypt with dek: create cipher: %w", err)
	}

	dekGCM, err := cipher.NewGCMWithRandomNonce(dekCipher)
	if err != nil {
		return "", fmt.Errorf("decrypt with dek: create gcm: %w", err)
	}

	raw, err := dekGCM.Open(nil, nil, eCipher, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt with dek: gcm open: %w", err)
	}

	return string(raw), nil
}

func decryptPtrWithDEK(ciphertext *string, rawDEK []byte) (*string, error) {
	if ciphertext == nil {
		return nil, nil
	}

	plain, err := decryptWithDEK(*ciphertext, rawDEK)
	if err != nil {
		return nil, fmt.Errorf("decrypt ptr with dek: %w", err)
	}

	return &plain, nil
}
