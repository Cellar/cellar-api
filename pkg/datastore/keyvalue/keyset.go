package keyvalue

import "fmt"

type KeySet struct {
	id string
}

func NewKeySet(id string) *KeySet {
	return &KeySet{id: id}
}

func (key KeySet) AccessLimit() string {
	return key.buildKey("accesslimit")
}

func (key KeySet) Access() string {
	return key.buildKey("access")
}

func (key KeySet) ContentType() string {
	return key.buildKey("contenttype")
}

func (key KeySet) Content() string {
	return key.buildKey("content")
}

func (key KeySet) ExpirationEpoch() string {
	return key.buildKey("expirationepoch")
}

func (key KeySet) Filename() string {
	return key.buildKey("filename")
}

func (key KeySet) AllKeys() []string {
	return []string{
		key.ContentType(),
		key.Content(),
		key.Access(),
		key.AccessLimit(),
		key.ExpirationEpoch(),
		key.Filename(),
	}
}

func (key KeySet) buildKey(tail string) string {
	return fmt.Sprintf("secrets:%s:%s", key.id, tail)
}
