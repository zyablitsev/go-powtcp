package proto

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// ResponseChallenge struct holds nonce and hash values for signed requested challenge data.
type ResponseChallenge struct {
	Nonce    uint64
	Hash     []byte
	Envelope *Envelope
	b        []byte
}

// Bytes returns gob-encoded struct representation.
func (s *ResponseChallenge) Bytes() ([]byte, error) {
	if s.b == nil {
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(s)
		if err != nil {
			err = fmt.Errorf("ResponseChallenge.Bytes: encode error: %w", err)
			return nil, err
		}
		s.b = buf.Bytes()
	}

	return s.b[:], nil
}
