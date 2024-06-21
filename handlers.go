package powtcp

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"time"

	"powtcp/pkg/pow"
	"powtcp/proto"
)

func (s *Service) handleRequestService(conn net.Conn) error {
	// prepare challenge data
	value := s.buffer.Pop()
	if value == nil {
		err := errors.New(
			"Service.handleRequestService: no available challenges")
		return err
	}

	s.mux.Lock()

	difficulty := 0
	difficulty = s.difficulty
	if s.pool.Len() > s.cfg.rpsTarget || s.rps > s.cfg.rpsTarget {
		difficulty = pow.MaxDifficulty
	}

	request := proto.RequestChallenge{
		Difficulty: difficulty,
		Value:      value,
	}

	// register challenge in the registry pool
	expires := time.Time{}
	b, err := request.Bytes()
	if err == nil {
		key := base64.StdEncoding.EncodeToString(b)
		expires = s.pool.Set(key)
	}

	s.mux.Unlock()

	if err != nil {
		err = fmt.Errorf("Service.handleRequestService: %w", err)
		return err
	}

	// sign challenge
	envelope, err := proto.SignRequestChallenge(request, expires, s.cfg.secret)
	if err != nil {
		err = fmt.Errorf("Service.handleRequestService: %w", err)
		return err
	}

	// create protocol message
	message := proto.Message{
		MessageType: proto.RequestChallengeType,
		Data:        &envelope,
	}

	// write protocol message with request challenge to the network
	enc := gob.NewEncoder(conn)
	err = enc.Encode(message)
	if err != nil {
		err = fmt.Errorf("Service.handleRequestService: %w", err)
		return err
	}

	return nil
}

func (s *Service) handleResponseChallenge(conn net.Conn, i interface{}) error {
	response, ok := i.(*proto.ResponseChallenge)
	if !ok {
		return errors.New("Service.handleResponseChallenge: bad request")
	}

	// check signature
	envelope := response.Envelope
	ok, err := envelope.Check(s.cfg.secret)
	if err != nil {
		err = fmt.Errorf("Service.handleResponseChallenge: %w", err)
		return err
	}
	if !ok {
		return errors.New("Service.handleResponseChallenge: bad signature")
	}

	// check proof
	request := envelope.RequestChallenge
	b, err := request.Bytes()
	if err != nil {
		err = fmt.Errorf("Service.handleResponseChallenge: %w", err)
		return err
	}
	ok = pow.Check(
		request.Value, request.Difficulty, response.Hash, response.Nonce)
	if !ok {
		return errors.New("Service.handleResponseChallenge: bad proof")
	}

	// adjust difficulty by time spent to resolve the challenge
	s.adjustDifficulty(response.Envelope.Expires)

	// check if expired
	if time.Now().After(response.Envelope.Expires) {
		return errors.New("Service.handleResponseChallenge: challenge expired")
	}

	// check challenge in the registry pool to prevent replay-attacks
	key := base64.StdEncoding.EncodeToString(b)
	ok = s.pool.Get(key)
	if !ok {
		return errors.New("Service.handleResponseChallenge: bad challenge")
	}

	// create protocol message with quote
	quote := quotes[response.Nonce%uint64(len(quotes))]
	message := proto.Message{
		MessageType: proto.ResponseServiceType,
		Data:        []byte(quote),
	}

	s.mux.Lock()
	s.rps++
	s.mux.Unlock()

	// write protocol message with quote to the network
	enc := gob.NewEncoder(conn)
	err = enc.Encode(message)
	if err != nil {
		err = fmt.Errorf("Service.handleResponseChallenge: %w", err)
		return err
	}

	return nil
}

func (s *Service) handleError(conn net.Conn, msg string) error {
	message := proto.Message{
		MessageType: proto.ErrorType,
		Data:        []byte(msg),
	}

	// write protocol message with error to the network
	enc := gob.NewEncoder(conn)
	err := enc.Encode(message)
	if err != nil {
		err = fmt.Errorf("Service.handleError: %w", err)
		return err
	}

	return nil
}
