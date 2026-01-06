package settings

import (
	"github.com/spf13/viper"
)

const (
	rateLimitKey                             = "rate_limit."
	rateLimitEnabledKey                      = rateLimitKey + "enabled"
	rateLimitWindowSecondsKey                = rateLimitKey + "window_seconds"
	rateLimitTier1RequestsPerWindowKey       = rateLimitKey + "tier1_requests_per_window"
	rateLimitTier2RequestsPerWindowKey       = rateLimitKey + "tier2_requests_per_window"
	rateLimitTier3RequestsPerWindowKey       = rateLimitKey + "tier3_requests_per_window"
	rateLimitHealthCheckRequestsPerWindowKey = rateLimitKey + "health_check_requests_per_window"
)

//go:generate mockgen -destination=../mocks/mock_ratelimit_configuration.go -package=mocks cellar/pkg/settings IRateLimitConfiguration
type IRateLimitConfiguration interface {
	Enabled() bool
	WindowSeconds() int
	Tier1RequestsPerWindow() int
	Tier2RequestsPerWindow() int
	Tier3RequestsPerWindow() int
	HealthCheckRequestsPerWindow() int
}

type RateLimitConfiguration struct{}

func NewRateLimitConfiguration() *RateLimitConfiguration {
	viper.SetDefault(rateLimitEnabledKey, true)
	viper.SetDefault(rateLimitWindowSecondsKey, 60)
	viper.SetDefault(rateLimitTier1RequestsPerWindowKey, 10)
	viper.SetDefault(rateLimitTier2RequestsPerWindowKey, 30)
	viper.SetDefault(rateLimitTier3RequestsPerWindowKey, 60)
	viper.SetDefault(rateLimitHealthCheckRequestsPerWindowKey, 120)
	return &RateLimitConfiguration{}
}

func (rlc RateLimitConfiguration) Enabled() bool {
	return viper.GetBool(rateLimitEnabledKey)
}

func (rlc RateLimitConfiguration) WindowSeconds() int {
	value := viper.GetInt(rateLimitWindowSecondsKey)
	if value < 1 {
		return 1
	}
	return value
}

func (rlc RateLimitConfiguration) Tier1RequestsPerWindow() int {
	value := viper.GetInt(rateLimitTier1RequestsPerWindowKey)
	if value < 1 {
		return 1
	}
	return value
}

func (rlc RateLimitConfiguration) Tier2RequestsPerWindow() int {
	value := viper.GetInt(rateLimitTier2RequestsPerWindowKey)
	if value < 1 {
		return 1
	}
	return value
}

func (rlc RateLimitConfiguration) Tier3RequestsPerWindow() int {
	value := viper.GetInt(rateLimitTier3RequestsPerWindowKey)
	if value < 1 {
		return 1
	}
	return value
}

func (rlc RateLimitConfiguration) HealthCheckRequestsPerWindow() int {
	value := viper.GetInt(rateLimitHealthCheckRequestsPerWindowKey)
	if value < 1 {
		return 1
	}
	return value
}
