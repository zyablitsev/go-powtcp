package proto

import (
	"encoding/gob"
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
}
