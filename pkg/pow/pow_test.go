package pow

import (
	"os"
	"testing"

	"powtcp/pkg/log"
)

func TestPowChallenge(t *testing.T) {
	log, _ := log.New("error", os.Stdout, os.Stderr)
	challenge := []byte("challenge")
	difficulty := 5
	hash, nonce := Fulfil(log, challenge, difficulty)
	ok := Check(challenge, difficulty, hash, nonce)
	if !ok {
		t.Error("check should return true value")
	}
}
