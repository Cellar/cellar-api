package redis_test

import (
	"cellar/pkg/datastore/redis"
	"cellar/testing/testhelpers"
	"fmt"
	"testing"
)

var id = "1234567890"
var sut = redis.NewRedisKeySet(id)

var keys = struct {
	access          string
	content         string
	accessLimit     string
	expirationEpoch string
}{
	access:          fmt.Sprintf("secrets:%s:access", id),
	content:         fmt.Sprintf("secrets:%s:content", id),
	accessLimit:     fmt.Sprintf("secrets:%s:accesslimit", id),
	expirationEpoch: fmt.Sprintf("secrets:%s:expirationepoch", id),
}

func TestRedisKey_Access(t *testing.T) {
	testhelpers.Equals(t, keys.access, sut.Access())
}

func TestRedisKey_Content(t *testing.T) {
	testhelpers.Equals(t, keys.content, sut.Content())
}

func TestRedisKey_AccessLimit(t *testing.T) {
	testhelpers.Equals(t, keys.accessLimit, sut.AccessLimit())
}

func TestRedisKey_ExpirationEpoch(t *testing.T) {
	testhelpers.Equals(t, keys.expirationEpoch, sut.ExpirationEpoch())
}

func TestRedisKey_AllKeys(t *testing.T) {
	for _, expected := range []string{keys.content, keys.access, keys.accessLimit, keys.expirationEpoch} {
		found := false
		for _, actual := range sut.AllKeys() {
			if actual == expected {
				found = true
			}
		}
		if !found {
			t.Fatalf("unable to find expected key '%s' in all keys", expected)
		}
	}
}
