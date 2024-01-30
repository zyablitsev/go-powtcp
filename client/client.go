package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"powtcp/pkg/challenge"
	"powtcp/pkg/log"
	"powtcp/pkg/pow"
	"powtcp/proto"
)

// Miner represents the pow tcp client.
type Miner struct {
	ctx  context.Context
	cfg  config
	log  logger
	conn net.Conn
}

// Init creates Miner instance.
func Init(ctx context.Context) (*Miner, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("can't load configuration: %v", err)
	}

	log, err := log.New(cfg.logLevel, os.Stdout, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("can't init logger: %v", err)
	}

	conn, err := net.Dial("tcp", cfg.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("dial tcp server failed: %v", err)
	}

	if cfg.connReadTTL > 0 {
		conn.SetReadDeadline(time.Now().Add(cfg.connReadTTL))
	}
	if cfg.connWriteTTL > 0 {
		conn.SetWriteDeadline(time.Now().Add(cfg.connWriteTTL))
	}

	m := &Miner{
		ctx:  ctx,
		cfg:  cfg,
		log:  log,
		conn: conn,
	}

	return m, nil
}

// Run starts performing requests until ctx is canceled.
func (m *Miner) Run() {
	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		for {
			err := m.GetQuote()
			if err != nil {
				m.log.Error(err)
				return
			}
		}
	}()

	// handle shutdown
	select {
	case <-ctx.Done():
	case <-m.ctx.Done():
		m.log.Info("interrupt syscall received")
		cancel()
	}

	m.log.Info("cleaning up…")
	m.cleanup()

	m.log.Info("waiting workers to stop…")
	wg.Wait()
	m.log.Info("cleanup done, shutdown")
}

// GetQuote requests service, fulfills challenge and write the quote to log.
func (m *Miner) GetQuote() error {
	// request service
	b := []byte{byte(proto.RequestService)}
	_, err := m.conn.Write(b)
	if err != nil {
		return err
	}

	// read for challenge request
	buf := make([]byte, 512)
	n, err := m.conn.Read(buf)
	if err != nil {
		return err
	}

	msgtype := int(buf[0])
	if msgtype != proto.RequestChallenge {
		return errors.New("wrong message type received")
	}

	v := buf[1:n]
	envelope, err := challenge.ParseEnvelope(v)
	if err != nil {
		return err
	}

	// fulfil the challenge
	hash, nonce := pow.Fulfil(m.log, envelope.Challenge(), envelope.Difficulty())

	// response with challenge proof
	b = make([]byte, 1+8+len(hash)+len(v)) // 1 byte for difficulty, 8 for nonce
	b[0] = byte(proto.ResponseChallenge)
	binary.BigEndian.PutUint64(b[1:9], nonce)
	copy(b[9:41], hash)
	copy(b[41:], v)

	_, err = m.conn.Write(b)
	if err != nil {
		return err
	}

	// read service response
	buf = make([]byte, 512)
	n, err = m.conn.Read(buf)
	if err != nil {
		return err
	}

	msgtype = int(buf[0])

	switch msgtype {
	case proto.ResponseService:
		m.log.Infof("got quote: %q", string(buf[1:n-1]))
	case proto.Error:
		return errors.New(string(buf[1:]))
	default:
		return errors.New("unknown message type")
	}

	return nil
}

func (m *Miner) cleanup() {
	m.conn.Close()
}
