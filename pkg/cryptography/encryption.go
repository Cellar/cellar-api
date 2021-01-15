package cryptography

import "cellar/pkg/models"

var Key = "CRYPTOGRAPHY"

//go:generate mockgen -destination=../mocks/mock_encryption.go -package=mocks . Encryption
type Encryption interface {
	Health() models.Health
	Encrypt(content []byte) (encryptedContent string, err error)
	Decrypt(content string) (decryptedContent []byte, err error)
}
