package proto

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
	"time"

	"powtcp/pkg/buffer"
	"powtcp/pkg/log"
)

func TestEnvelope(t *testing.T) {
	log, _ := log.New("error", os.Stdout, os.Stderr)
	challengeLen := 8
	bufferSize := 10
	buf, err := buffer.New(log, challengeLen, bufferSize)
	if err != nil {
		t.Fatal(err)
	}

	ttl := time.Second
	secret := "secret"
	difficulty := 5

	for i := 0; i < bufferSize*5; i++ {
		value := buf.Pop()
		expires := time.Now().Add(ttl)
		request := RequestChallenge{
			Difficulty: difficulty,
			Value:      value,
		}

		envelope, err := SignRequestChallenge(request, expires, secret)
		if err != nil {
			t.Fatal(err)
		}

		b, err := envelope.Bytes()
		if err != nil {
			t.Fatal(err)
		}

		buf := bytes.NewBuffer(b)
		dec := gob.NewDecoder(buf)
		envelope = Envelope{}
		err = dec.Decode(&envelope)
		if err != nil {
			t.Fatal(err)
		}

		ok, err := envelope.Check(secret)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("bad signature")
		}
	}
}
