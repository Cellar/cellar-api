package models

type ConfigResponse struct {
	Limits LimitsConfig `json:"limits"`
}

type LimitsConfig struct {
	MaxFileSizeMB        int `json:"maxFileSizeMB" example:"8"`
	MaxAccessCount       int `json:"maxAccessCount" example:"100"`
	MaxExpirationSeconds int `json:"maxExpirationSeconds" example:"604800"`
}
