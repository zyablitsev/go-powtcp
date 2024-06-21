package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"powtcp/pkg/pow"
	"powtcp/proto"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("main: can't load configuration: %w", err))
	}

	conn, err := net.Dial("tcp", cfg.serverAddr)
	if err != nil {
		log.Fatal(fmt.Errorf("main: dial tcp server failed: %w", err))
	}
	defer conn.Close()

	if cfg.connReadTTL > 0 {
		err = conn.SetReadDeadline(time.Now().Add(cfg.connReadTTL))
		if err != nil {
			err = fmt.Errorf("main: can't set read deadline: %w", err)
			log.Fatal(err)
		}
	}
	if cfg.connWriteTTL > 0 {
		err = conn.SetWriteDeadline(time.Now().Add(cfg.connWriteTTL))
		if err != nil {
			err = fmt.Errorf("main: can't set write deadline: %w", err)
			log.Fatal(err)
		}
	}

	for {
		quote, err := getQuote(conn)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if errors.Is(err, syscall.EPIPE) {
				return
			}
			err = fmt.Errorf("main loop: %w", err)
			log.Println(err)
		} else {
			log.Println(quote)
		}
	}
}

// getQuote requests service, fulfills challenge and write the quote to stdout.
func getQuote(conn net.Conn) (string, error) {
	// write request service
	message := proto.Message{
		MessageType: proto.RequestServiceType,
	}
	enc := gob.NewEncoder(conn)
	err := enc.Encode(message)
	if err != nil {
		err = fmt.Errorf("getQuote: encode message error: %w", err)
		return "", err
	}

	// read challenge request
	message = proto.Message{}
	dec := gob.NewDecoder(conn)
	err = dec.Decode(&message)
	if err != nil {
		err = fmt.Errorf("getQuote: decode message error: %w", err)
		return "", err
	}

	if message.MessageType == proto.ErrorType {
		err := errors.New("getQuote: request service failed")
		description, ok := message.Data.([]byte)
		if ok {
			err = fmt.Errorf(
				"getQuote: request service failed: %s", string(description))
		}
		return "", err
	}
	if message.MessageType != proto.RequestChallengeType {
		return "", errors.New("getQuote: wrong message type received")
	}

	envelope, ok := message.Data.(*proto.Envelope)
	if !ok {
		return "", errors.New("getQuote: bad data for request challenge message")
	}

	// fulfil the challenge
	request := envelope.RequestChallenge
	hash, nonce := pow.Fulfil(
		request.Value, request.Difficulty, envelope.Expires)
	if hash == nil {
		err := fmt.Errorf(
			"getQuote: can't fulfil the challenge with difficulty %d in time",
			request.Difficulty)
		return "", err
	}

	// write challenge proof
	response := proto.ResponseChallenge{
		Nonce:    nonce,
		Hash:     hash,
		Envelope: envelope,
	}
	message = proto.Message{
		MessageType: proto.ResponseChallengeType,
		Data:        &response,
	}
	enc = gob.NewEncoder(conn)
	err = enc.Encode(message)
	if err != nil {
		err = fmt.Errorf("getQuote: encode message error: %w", err)
		return "", err
	}

	// read service response
	message = proto.Message{}
	dec = gob.NewDecoder(conn)
	err = dec.Decode(&message)
	if err != nil {
		err = fmt.Errorf("getQuote: decode message error: %w", err)
		return "", err
	}

	if message.MessageType == proto.ErrorType {
		err := errors.New("getQuote: service response failed")
		description, ok := message.Data.([]byte)
		if ok {
			err = fmt.Errorf(
				"getQuote: service response failed: %s", string(description))
		}
		return "", err
	}
	if message.MessageType != proto.ResponseServiceType {
		return "", errors.New("getQuote: wrong message type received")
	}

	quote, ok := message.Data.([]byte)
	if !ok {
		return "", errors.New("getQuote: bad data for response service message")
	}

	return string(quote), nil
}

const (
	defaultConnReadTTL  = "1s"
	defaultConnWriteTTL = "1s"
)

type config struct {
	serverAddr   string
	connReadTTL  time.Duration
	connWriteTTL time.Duration
}

func loadConfig() (config, error) {
	serverAddrEnv := os.Getenv("CLIENT_SERVER_ADDR")
	if serverAddrEnv == "" {
		err := fmt.Errorf(
			"loadConfig: bad CLIENT_SERVER_ADDR env value %q", serverAddrEnv)
		return config{}, err
	}

	connReadTTLEnv := os.Getenv("CLIENT_CONNREAD_TTL")
	if connReadTTLEnv == "" {
		connReadTTLEnv = defaultConnReadTTL
	}
	connReadTTL, err := time.ParseDuration(connReadTTLEnv)
	if err != nil {
		err = fmt.Errorf(
			"loadConfig: bad CLIENT_CONNREAD_TTL env value %q: %w",
			connReadTTLEnv, err)
		return config{}, err
	}

	connWriteTTLEnv := os.Getenv("CLIENT_CONNWRITE_TTL")
	if connWriteTTLEnv == "" {
		connWriteTTLEnv = defaultConnWriteTTL
	}
	connWriteTTL, err := time.ParseDuration(connWriteTTLEnv)
	if err != nil {
		err = fmt.Errorf(
			"loadConfig: bad CLIENT_CONNWRITE_TTL env value %q: %w",
			connWriteTTLEnv, err)
		return config{}, err
	}

	cfg := config{
		serverAddr:   serverAddrEnv,
		connReadTTL:  connReadTTL,
		connWriteTTL: connWriteTTL,
	}

	return cfg, nil
}
