package redis_test

import (
	"cellar/pkg/datastore/redis"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var id = "1234567890"
var sut = redis.NewRedisKeySet(id)

var keys = struct {
	access          string
	contentType     string
	content         string
	accessLimit     string
	expirationEpoch string
}{
	access:          fmt.Sprintf("secrets:%s:access", id),
	contentType:     fmt.Sprintf("secrets:%s:contenttype", id),
	content:         fmt.Sprintf("secrets:%s:content", id),
	accessLimit:     fmt.Sprintf("secrets:%s:accesslimit", id),
	expirationEpoch: fmt.Sprintf("secrets:%s:expirationepoch", id),
}

func TestRedisKey_Access(t *testing.T) {
	assert.Equal(t, keys.access, sut.Access())
}

func TestRedisKey_ContentType(t *testing.T) {
	assert.Equal(t, keys.contentType, sut.ContentType())
}

func TestRedisKey_Content(t *testing.T) {
	assert.Equal(t, keys.content, sut.Content())
}

func TestRedisKey_AccessLimit(t *testing.T) {
	assert.Equal(t, keys.accessLimit, sut.AccessLimit())
}

func TestRedisKey_ExpirationEpoch(t *testing.T) {
	assert.Equal(t, keys.expirationEpoch, sut.ExpirationEpoch())
}

func TestRedisKey_AllKeys(t *testing.T) {
	allKeys := sut.AllKeys()
	for _, expected := range []string{keys.contentType, keys.content, keys.access, keys.accessLimit, keys.expirationEpoch} {
		assert.Contains(t, allKeys, expected)
	}
}
