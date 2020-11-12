package datastore

import "fmt"

type RedisKey struct {
	id string
}

func NewRedisKeySet(id string) *RedisKey {
	return &RedisKey{id: id}
}

func (key RedisKey) AccessLimit() string {
	return key.buildKey("accesslimit")()
}

func (key RedisKey) Access() string {
	return key.buildKey("access")()
}

func (key RedisKey) Content() string {
	return key.buildKey("content")()
}

func (key RedisKey) ExpirationEpoch() string {
	return key.buildKey("expirationepoch")()
}

func (key RedisKey) AllKeys() []string {
	return []string{
		key.Content(),
		key.Access(),
		key.AccessLimit(),
		key.ExpirationEpoch(),
	}
}

func (key RedisKey) buildKey(tail string) func() string {
	return func() string {
		return fmt.Sprintf("secrets:%s:%s", key.id, tail)
	}
}
