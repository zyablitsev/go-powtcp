package server

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"powtcp/pkg/challenge"
	"powtcp/pkg/log"
	"powtcp/pkg/pow"
	"powtcp/proto"
)

// Service represents the pow tcp server.
type Service struct {
	ctx context.Context
	cfg config
	log logger

	quotes     *quotes
	buffer     *challenge.Buffer
	registry   *challenge.RegistryPool
	difficulty int
	rps        int
	mux        sync.RWMutex
	listener   net.Listener
}

// Init creates Service instance.
func Init(ctx context.Context) (*Service, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("can't load configuration: %v", err)
	}

	log, err := log.New(cfg.logLevel, os.Stdout, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("can't init logger: %v", err)
	}

	buffer, err := challenge.NewBuffer(
		log,
		cfg.challengeLen,
		cfg.challengeBufferSize)
	if err != nil {
		return nil, fmt.Errorf("can't init challenge buffer: %v", err)
	}

	registry := challenge.NewRegistryPool(
		ctx,
		cfg.challengeTTL,
		cfg.challengePoolCleanupInterval)

	quotes := newQuotes()

	s := &Service{
		ctx: ctx,
		cfg: cfg,
		log: log,

		quotes:     quotes,
		buffer:     buffer,
		registry:   registry,
		difficulty: cfg.difficulty,
	}

	return s, nil
}

// Run begins to listen tcp socket until ctx is canceled.
func (s *Service) Run() {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", s.cfg.ip, s.cfg.port)
	lc := &net.ListenConfig{}
	var err error
	s.listener, err = lc.Listen(ctx, "tcp", addr)
	if err != nil {
		s.log.Errorf("tcp socket bind failed: %v", err)
		return
	}
	s.log.Infof("serving tcp socket on %q…", addr)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		for {
			conn, err := s.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				s.log.Error(err)
				continue
			}

			go s.handleConn(conn)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		t := time.NewTicker(time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.mux.Lock()
				if s.rps < s.cfg.rpsTarget && s.difficulty > 1 {
					s.difficulty -= 1
				} else if s.rps > s.cfg.rpsTarget &&
					s.difficulty < math.MaxInt64 {

					s.difficulty += 1
				}
				s.rps = 0
				s.mux.Unlock()
			}
		}
	}()

	// handle shutdown
	select {
	case <-ctx.Done():
	case <-s.ctx.Done():
		s.log.Info("interrupt syscall received")
		cancel()
	}

	s.log.Info("cleaning up…")
	s.cleanup()

	s.log.Info("waiting workers to stop…")
	wg.Wait()
	s.log.Info("cleanup done, shutdown")
}

func (s *Service) handleConn(conn net.Conn) {
	defer conn.Close()

	if s.cfg.connReadTTL > 0 {
		conn.SetReadDeadline(time.Now().Add(s.cfg.connReadTTL))
	}
	if s.cfg.connWriteTTL > 0 {
		conn.SetWriteDeadline(time.Now().Add(s.cfg.connWriteTTL))
	}

	s.log.Infof("accept client conn %q", conn.RemoteAddr().String())
	defer func() {
		s.log.Infof("close client conn %q", conn.RemoteAddr().String())
	}()

	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			s.log.Error(err)
			return
		}
		if n == 0 {
			continue
		}

		msgtype := int(buf[0])

		switch msgtype {
		case proto.RequestService:
			s.log.Info("got request service")
			err = s.requestChallenge(conn)
			if err != nil {
				s.log.Error(err)
				return
			}
		case proto.ResponseChallenge:
			s.log.Info("got response challenge")
			err = s.handleService(conn, buf[1:n])
			if err != nil {
				s.log.Error(err)
				_ = s.handleError(conn, err.Error())
				return
			}
		default:
			s.log.Infof("got unknown message type: %d", msgtype)
			err = s.handleError(
				conn, fmt.Sprintf("unknown message type: %d", msgtype))
			if err != nil {
				s.log.Error(err)
			}
			return
		}
	}
}

func (s *Service) requestChallenge(conn net.Conn) error {
	v := s.buffer.Pop()

	b := make([]byte, 1+len(v)) // 1 byte for difficulty
	b[0] = byte(s.difficulty)
	copy(b[1:], v)

	key := base64.StdEncoding.EncodeToString(b)
	expires := s.registry.Set(key)

	envelope, err := challenge.NewEnvelope(
		s.difficulty, v, expires, s.cfg.secret)
	if err != nil {
		return err
	}

	signed := envelope.Signed()
	b = make([]byte, 1+len(signed))
	b[0] = byte(proto.RequestChallenge)
	copy(b[1:], signed)

	_, err = conn.Write(b)

	return err
}

func (s *Service) handleService(conn net.Conn, buf []byte) error {
	if len(buf) < 40 {
		return errors.New("bad data")
	}

	// check proof
	nonceBytes := buf[:8]
	hash := buf[8:40]

	envelope, err := challenge.ParseEnvelope(buf[40:])
	if err != nil {
		return err
	}

	ok := envelope.Validate(s.cfg.secret)
	if !ok {
		return errors.New("bad envelope")
	}

	challenge := envelope.Challenge()
	difficulty := envelope.Difficulty()
	nonce := binary.BigEndian.Uint64(nonceBytes)
	ok = pow.Check(challenge, difficulty, hash, nonce)
	if !ok {
		return errors.New("bad proof")
	}

	b := make([]byte, 1+len(challenge)) // 1 byte for difficulty
	b[0] = byte(difficulty)
	copy(b[1:], challenge)

	key := base64.StdEncoding.EncodeToString(b)
	ok = s.registry.Get(key)
	if !ok {
		return errors.New("bad challenge")
	}

	// get quote
	quote := []byte(s.quotes.get())
	b = make([]byte, 1+len(quote))
	b[0] = byte(proto.ResponseService)
	copy(b[1:], quote)

	s.mux.Lock()
	s.rps++
	s.mux.Unlock()

	_, err = conn.Write(b)

	return err
}

func (s *Service) handleError(conn net.Conn, msg string) error {
	b := make([]byte, 1+len(msg))
	b[0] = byte(proto.Error)
	copy(b[1:], []byte(msg))

	_, err := conn.Write(b)

	return err
}

func (s *Service) cleanup() {
	s.listener.Close()
}
