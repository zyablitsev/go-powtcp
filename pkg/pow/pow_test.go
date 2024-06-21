package pow

import (
	"testing"
	"time"
)

func TestPowChallenge(t *testing.T) {
	challenge := []byte("challenge")
	difficulty := 5
	ttl := time.Second
	expires := time.Now().Add(ttl)
	hash, nonce := Fulfil(challenge, difficulty, expires)
	ok := Check(challenge, difficulty, hash, nonce)
	if !ok {
		t.Error("check should return true value")
	}
}
