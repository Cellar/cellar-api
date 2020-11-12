package settings

import (
	"github.com/spf13/viper"
)

type IRedisConfiguration interface {
	Host() string
	Port() int
	Password() string
	DB() int
}

const (
	redisKey         = "redis."
	redisHostKey     = redisKey + "host"
	redisPortKey     = redisKey + "port"
	redisPasswordKey = redisKey + "password"
	redisDBKey       = redisKey + "db"
)

type RedisConfiguration struct{}

func NewRedisConfiguration() *RedisConfiguration {
	viper.SetDefault(redisHostKey, "localhost")
	viper.SetDefault(redisPortKey, 6379)
	viper.SetDefault(redisPasswordKey, "")
	viper.SetDefault(redisDBKey, 0)
	return &RedisConfiguration{}
}

func (rds RedisConfiguration) Host() string {
	return viper.GetString(redisHostKey)
}

func (rds RedisConfiguration) Port() int {
	return viper.GetInt(redisPortKey)
}

func (rds RedisConfiguration) Password() string {
	return viper.GetString(redisPasswordKey)
}

func (rds RedisConfiguration) DB() int {
	return viper.GetInt(redisDBKey)
}
