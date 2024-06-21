package proto

import (
	"bytes"
	"crypto/subtle"
	"encoding/gob"
	"fmt"
	"time"

	"golang.org/x/crypto/sha3"
)

// Envelope struct represents signed RequestChallenge envelope.
type Envelope struct {
	RequestChallenge RequestChallenge
	Expires          time.Time
	Signature        []byte
	b                []byte
}

// SignRequestChallenge constructor.
func SignRequestChallenge(
	requestChallenge RequestChallenge,
	expires time.Time,
	secret string,
) (Envelope, error) {
	requestBytes, err := requestChallenge.Bytes()
	if err != nil {
		err = fmt.Errorf("SignRequestChallenge: %w", err)
		return Envelope{}, err
	}
	expiresBytes, err := expires.MarshalBinary()
	if err != nil {
		err = fmt.Errorf(
			"SignRequestChallenge: marshal expires ts error: %w", err)
		return Envelope{}, err
	}

	signature := sign(append(requestBytes, expiresBytes...), []byte(secret))

	s := Envelope{
		RequestChallenge: requestChallenge,
		Expires:          expires,
		Signature:        signature,
	}

	_, err = s.Bytes()
	if err != nil {
		err = fmt.Errorf("SignRequestChallenge: %w", err)
		return Envelope{}, err
	}

	return s, nil
}

// Bytes returns gob-encoded struct representation.
func (s *Envelope) Bytes() ([]byte, error) {
	if s.b == nil {
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(s)
		if err != nil {
			err = fmt.Errorf("Envelope.Bytes: encode error: %w", err)
			return nil, err
		}
		s.b = buf.Bytes()
	}

	return s.b[:], nil
}

// Check signature.
func (s *Envelope) Check(secret string) (bool, error) {
	requestBytes, err := s.RequestChallenge.Bytes()
	if err != nil {
		err = fmt.Errorf("Envelope.Check: %w", err)
		return false, err
	}

	expiresBytes, err := s.Expires.MarshalBinary()
	if err != nil {
		err = fmt.Errorf("Envelope.Check: marshal expires ts error: %w", err)
		return false, err
	}

	signature := sign(append(requestBytes, expiresBytes...), []byte(secret))

	return subtle.ConstantTimeCompare(s.Signature, signature) == 1, nil
}

// sign signs data using key by calculating sha3-512 hash
// of their concatenation.
func sign(data, key []byte) []byte {
	b := make([]byte, len(data)+len(key))
	idx := copy(b, data)
	copy(b[idx:], key)
	h := sha3.Sum512(b)
	return h[:]
}
