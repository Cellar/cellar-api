package redis

import "fmt"

type RedisKey struct {
	id string
}

func NewRedisKeySet(id string) *RedisKey {
	return &RedisKey{id: id}
}

func (key RedisKey) AccessLimit() string {
	return key.buildKey("accesslimit")
}

func (key RedisKey) Access() string {
	return key.buildKey("access")
}

func (key RedisKey) ContentType() string {
	return key.buildKey("contenttype")
}

func (key RedisKey) Content() string {
	return key.buildKey("content")
}

func (key RedisKey) ExpirationEpoch() string {
	return key.buildKey("expirationepoch")
}

func (key RedisKey) Filename() string {
	return key.buildKey("filename")
}

func (key RedisKey) AllKeys() []string {
	return []string{
		key.ContentType(),
		key.Content(),
		key.Access(),
		key.AccessLimit(),
		key.ExpirationEpoch(),
		key.Filename(),
	}
}

func (key RedisKey) buildKey(tail string) string {
	return fmt.Sprintf("secrets:%s:%s", key.id, tail)
}
