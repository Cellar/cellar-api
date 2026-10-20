package keyvalue_test

import (
	"cellar/pkg/datastore/keyvalue"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var id = "1234567890"
var sut = keyvalue.NewKeySet(id)

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

func TestKeySet_Access(t *testing.T) {
	assert.Equal(t, keys.access, sut.Access())
}

func TestKeySet_ContentType(t *testing.T) {
	assert.Equal(t, keys.contentType, sut.ContentType())
}

func TestKeySet_Content(t *testing.T) {
	assert.Equal(t, keys.content, sut.Content())
}

func TestKeySet_AccessLimit(t *testing.T) {
	assert.Equal(t, keys.accessLimit, sut.AccessLimit())
}

func TestKeySet_ExpirationEpoch(t *testing.T) {
	assert.Equal(t, keys.expirationEpoch, sut.ExpirationEpoch())
}

func TestKeySet_AllKeys(t *testing.T) {
	allKeys := sut.AllKeys()
	for _, expected := range []string{keys.contentType, keys.content, keys.access, keys.accessLimit, keys.expirationEpoch} {
		assert.Contains(t, allKeys, expected)
	}
}
