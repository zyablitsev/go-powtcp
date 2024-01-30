package challenge

import (
	"os"
	"powtcp/pkg/log"
	"testing"
	"time"
)

func TestEnvelope(t *testing.T) {
	log, _ := log.New("error", os.Stdout, os.Stderr)
	challengeLen := 8
	bufferSize := 10
	buf, err := NewBuffer(log, challengeLen, bufferSize)
	if err != nil {
		t.Fatal(err)
	}

	ttl := time.Second
	secret := "secret"
	difficulty := 5

	for i := 0; i < bufferSize*5; i++ {
		challenge := buf.Pop()
		expires := time.Now().Add(ttl)
		envelope, err := NewEnvelope(difficulty, challenge, expires, secret)
		if err != nil {
			t.Error(err)
		}

		envelope, err = ParseEnvelope(envelope.signed)
		if err != nil {
			t.Error(err)
		}

		ok := envelope.Validate(secret)
		if !ok {
			t.Error("bad envelope")
		}
	}
}
