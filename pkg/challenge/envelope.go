package challenge

import (
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"time"

	"golang.org/x/crypto/sha3"
)

const expiresLen = 8

// Envelope with signed challenge data.
type Envelope struct {
	difficulty int
	challenge  []byte
	expires    time.Time

	payload []byte

	signature []byte
	signed    []byte
}

// NewEnvelope constructor.
func NewEnvelope(
	difficulty int,
	challenge []byte,
	expires time.Time,
	secret string,
) (Envelope, error) {
	// validate input
	if len(secret) < 1 {
		return Envelope{}, errors.New("secret shouldn't be empty")
	}
	if len(challenge) > 255 {
		return Envelope{}, errors.New("challenge should be shorter than 256")
	}
	if difficulty > 255 {
		return Envelope{}, errors.New("difficulty should be lower than 256")
	}

	challengeLen := len(challenge)

	// 1 byte for difficulty
	// 1 byte for challengeLen
	payloadBytes := make([]byte, expiresLen+1+1+challengeLen)

	// write expires bytes to payloadBytes
	binary.BigEndian.PutUint64(payloadBytes[:expiresLen], uint64(expires.UnixNano()))

	// copy challenge bytes to payloadBytes
	idx := expiresLen
	payloadBytes[idx] = byte(difficulty)
	idx++
	payloadBytes[idx] = byte(challengeLen)
	idx++
	idx += copy(payloadBytes[idx:], challenge)

	// sign payload bytes with secret
	signature := sign(payloadBytes, []byte(secret))

	signatureLen := len(signature)
	signedBytes := make([]byte, len(payloadBytes)+1+signatureLen)
	idx = copy(signedBytes, payloadBytes)
	signedBytes[idx] = byte(signatureLen)
	idx++
	copy(signedBytes[idx:], signature)

	envelope := Envelope{
		difficulty: difficulty,
		challenge:  challenge,
		expires:    expires,

		payload: payloadBytes,

		signature: signature,
		signed:    signedBytes,
	}

	return envelope, nil
}

// ParseEnvelope Envelope from bytes.
func ParseEnvelope(b []byte) (Envelope, error) {
	if len(b) < expiresLen+2 {
		return Envelope{}, errors.New("bad data")
	}

	// parse expires ts
	expiresUnixNano := binary.BigEndian.Uint64(b[0:expiresLen])
	expires := time.Unix(0, int64(expiresUnixNano))
	if expires.Before(time.Now()) {
		return Envelope{}, errors.New("challenge expired")
	}

	// parse difficulty
	idx := expiresLen
	difficulty := int(b[idx])

	// parse challenge
	idx++
	challengeLen := int(b[idx])
	idx++
	if len(b) < (idx + challengeLen) {
		return Envelope{}, errors.New("wrong challenge format")
	}
	challenge := b[idx : idx+challengeLen]
	idx += challengeLen

	// parse signature
	if len(b) < (idx + 1) {
		return Envelope{}, errors.New("wrong signature format")
	}
	signatureLen := int(b[idx])
	idx++
	if len(b) < (idx + signatureLen) {
		return Envelope{}, errors.New("bad signature")
	}

	payloadBytes := b[:idx-1]
	signature := b[idx : idx+signatureLen]

	envelope := Envelope{
		difficulty: difficulty,
		challenge:  challenge,
		expires:    expires,

		payload: payloadBytes,

		signature: signature,
		signed:    b,
	}

	return envelope, nil
}

// Validate checks data signature.
func (s *Envelope) Validate(secret string) bool {
	calcSignature := sign(s.payload, []byte(secret))

	return subtle.ConstantTimeCompare(s.signature, calcSignature) == 1
}

// Difficulty returns difficulty value.
func (s *Envelope) Difficulty() int {
	return s.difficulty
}

// Challenge returns bytes of challenge data.
func (s *Envelope) Challenge() []byte {
	return s.challenge[:]
}

// Signed returns bytes of signed data.
func (s *Envelope) Signed() []byte {
	return s.signed[:]
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
