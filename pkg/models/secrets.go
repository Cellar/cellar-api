package models

import (
	"mime/multipart"
	"time"
)

type CreateSecretRequest struct {
	Content         *string `json:"content" example:"my very secret text"`
	AccessLimit     *int    `json:"access_limit" example:"10"`
	ExpirationEpoch *int64  `json:"expiration_epoch" example:"1577836800"`
}

type CreateSecretFileRequest struct {
	FileHeader      *multipart.FileHeader `json:"file"`
	Content         *string               `json:"content" example:"my very secret text"`
	AccessLimit     *int                  `json:"access_limit" example:"10"`
	ExpirationEpoch *int64                `json:"expiration_epoch" example:"1577836800"`
}

type SecretMetadataResponse struct {
	ID          string        `json:"id" example:"22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"`
	AccessCount int           `json:"access_count" example:"1"`
	AccessLimit int           `json:"access_limit" example:"10"`
	Expiration  FormattedTime `json:"expiration" swaggertype:"string" example:"1970-01-01 00:00:00 UTC"`
}

type Secret struct {
	ID              string
	Content         string
	AccessCount     int
	AccessLimit     int
	ExpirationEpoch int64
}

func NewSecret(id, content string, accessCount, accessLimit int, expirationEpoch int64) *Secret {
	return &Secret{
		ID:              id,
		Content:         content,
		AccessCount:     accessCount,
		AccessLimit:     accessLimit,
		ExpirationEpoch: expirationEpoch,
	}
}

type SecretContentResponse struct {
	ID      string `json:"id" example:"22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"`
	Content string `json:"content" example:"my very secret text"`
}

func (secret *Secret) Expiration() FormattedTime {
	return FormattedTime(time.Unix(secret.ExpirationEpoch, 0).UTC())
}

func (secret *Secret) Duration() time.Duration {
	return time.Until(secret.Expiration().Time().UTC())
}
