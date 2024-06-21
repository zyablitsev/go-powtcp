package buffer

import (
	"bytes"
	"os"
	"testing"

	"powtcp/pkg/log"
)

func TestChallengeBuffer(t *testing.T) {
	log, _ := log.New("error", os.Stdout, os.Stderr)
	buf, err := New(log, 8, 10)
	if err != nil {
		t.Fatal(err)
	}

	zero := make([]byte, 8)

	v1 := buf.Pop()
	if bytes.Compare(v1, zero) == 0 {
		t.Error("v1 shouldn't be zero value")
	}

	v2 := buf.Pop()
	if bytes.Compare(v1, zero) == 0 {
		t.Error("v2 shouldn't be zero value")
	}

	if bytes.Compare(v1, v2) == 0 {
		t.Error("v1 and v2 values should be different")
	}
}
