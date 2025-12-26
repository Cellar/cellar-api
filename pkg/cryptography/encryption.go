package cryptography

import (
	"cellar/pkg/models"
	"context"
)

var Key = "CRYPTOGRAPHY"

//go:generate mockgen -destination=../mocks/mock_encryption.go -package=mocks . Encryption
type Encryption interface {
	Health(ctx context.Context) models.Health
	Encrypt(ctx context.Context, plaintext []byte) (ciphertext string, err error)
	Decrypt(ctx context.Context, ciphertext string) (plaintext []byte, err error)
}
