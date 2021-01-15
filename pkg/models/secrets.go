package models

import (
	"time"
)

const (
	ContentTypeFile = "file"
	ContentTypeText = "text"
)

type (
	CreateSecretRequest struct {
		Content         *string `json:"content" example:"my very secret text"`
		AccessLimit     *int    `json:"access_limit" example:"10"`
		ExpirationEpoch *int64  `json:"expiration_epoch" example:"1577836800"`
	}

	SecretMetadataResponse struct {
		ID          string        `json:"id" example:"22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"`
		AccessCount int           `json:"access_count" example:"1"`
		AccessLimit int           `json:"access_limit" example:"10"`
		Expiration  FormattedTime `json:"expiration" swaggertype:"string" example:"1970-01-01 00:00:00 UTC"`
	}

	SecretMetadataResponseV2 struct {
		ID          string        `json:"id" example:"22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"`
		AccessCount int           `json:"access_count" example:"1"`
		AccessLimit int           `json:"access_limit" example:"10"`
		ContentType ContentType   `json:"content_type" swaggertype:"string" example:"text"`
		Expiration  FormattedTime `json:"expiration" swaggertype:"string" example:"1970-01-01 00:00:00 UTC"`
	}

	ContentType string

	Secret struct {
		ID              string
		Content         []byte
		CipherText      string
		ContentType     string
		AccessCount     int
		AccessLimit     int
		ExpirationEpoch int64
	}

	SecretMetadata struct {
		ID          string
		ContentType ContentType
		AccessCount int
		AccessLimit int
		Expiration  FormattedTime
	}

	SecretContentResponse struct {
		ID      string `json:"id" example:"22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"`
		Content string `json:"content" example:"my very secret text"`
	}
)

func (secret *Secret) Expiration() FormattedTime {
	return FormattedTime(time.Unix(secret.ExpirationEpoch, 0).UTC())
}

func (secret *Secret) Duration() time.Duration {
	return time.Until(secret.Expiration().Time().UTC())
}

func (secret *Secret) Metadata() *SecretMetadata {
	return &SecretMetadata{
		ID:          secret.ID,
		ContentType: ContentType(secret.ContentType),
		AccessCount: secret.AccessCount,
		AccessLimit: secret.AccessLimit,
		Expiration:  secret.Expiration(),
	}
}
