package config

type EncryptionConfig struct {
	PasswordKEK string `envconfig:"PASSWORD_KEK"`
}
