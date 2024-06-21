package proto

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// RequestChallenge struct represents challenge request.
type RequestChallenge struct {
	Difficulty int
	Value      []byte
	b          []byte
}

// Bytes returns gob-encoded struct representation.
func (s *RequestChallenge) Bytes() ([]byte, error) {
	if s.b == nil {
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(s)
		if err != nil {
			err = fmt.Errorf("RequestChallenge.Bytes: encode error: %w", err)
			return nil, err
		}
		s.b = buf.Bytes()
	}

	return s.b[:], nil
}
