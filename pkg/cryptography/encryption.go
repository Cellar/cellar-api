package cryptography

import "cellar/pkg/models"

var Key = "CRYPTOGRAPHY"

//go:generate mockgen -destination=../mocks/mock_encryption.go -package=mocks . Encryption
type Encryption interface {
	Health() models.Health
	Encrypt(plaintext []byte) (ciphertext string, err error)
	Decrypt(ciphertext string) (plaintext []byte, err error)
}
