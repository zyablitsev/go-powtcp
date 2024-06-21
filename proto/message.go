package proto

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func init() {
	// register gob interface types
	gob.Register((*Envelope)(nil))
	gob.Register((*ResponseChallenge)(nil))
}

// MessageType represents protocol message type as int.
type MessageType uint64

const (
	RequestServiceType MessageType = iota + 1
	RequestChallengeType
	ResponseChallengeType
	ResponseServiceType
	ErrorType
)

// Message struct represents protocol message.
type Message struct {
	MessageType MessageType
	Data        interface{}
	b           []byte
}

// Bytes returns gob-encoded struct representation.
func (s *Message) Bytes() ([]byte, error) {
	if s.b == nil {
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(s)
		if err != nil {
			err = fmt.Errorf("Message.Bytes: encode error: %w", err)
			return nil, err
		}
		s.b = buf.Bytes()
	}

	return s.b[:], nil
}
