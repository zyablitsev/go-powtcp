package powtcp

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"powtcp/pkg/buffer"
	"powtcp/pkg/log"
	"powtcp/pkg/pool"
	"powtcp/pkg/pow"
	"powtcp/proto"
)

var quotes = []string{
	"Don't be a Jack of all trades and master of none.",
	"Be sticky to your research goal.",
}

// Service represents the pow tcp server.
type Service struct {
	ctx context.Context
	cfg config
	log logger

	buffer     *buffer.Buffer
	pool       *pool.Pool
	difficulty int
	rps        int
	mux        sync.RWMutex
	listener   net.Listener
}

// Init creates Service instance.
func Init(ctx context.Context) (*Service, error) {
	cfg, err := loadConfig()
	if err != nil {
		err = fmt.Errorf("Service.Init: can't load configuration: %w", err)
		return nil, err
	}

	log, err := log.New(cfg.logLevel, os.Stdout, os.Stderr)
	if err != nil {
		err = fmt.Errorf("Service.Init: can't init logger: %w", err)
		return nil, err
	}

	// init challenge bucket buffer
	bufferSize := cfg.rpsTarget
	buffer, err := buffer.New(
		log,
		cfg.challengeLen,
		bufferSize)
	if err != nil {
		err = fmt.Errorf("Service.Init: can't init challenge buffer: %w", err)
		return nil, err
	}

	// init registry pool to track active challenges
	pool := pool.New(
		ctx,
		cfg.challengeTTL,
		cfg.challengePoolCleanupInterval)

	// init difficulty with min value on start
	difficulty := pow.MinDifficulty

	s := &Service{
		ctx: ctx,
		cfg: cfg,
		log: log,

		buffer:     buffer,
		pool:       pool,
		difficulty: difficulty,
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
		s.log.Errorf("Service.Run: tcp socket bind failed: %w", err)
		return
	}
	s.log.Infof("Service.Run: serving tcp socket on %q…", addr)

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
				err = fmt.Errorf(
					"Service.Run: error on s.listener.Accept: %w", err)
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
				// adjust difficulty by rps
				s.adjustDifficultyByRPS()
			}
		}
	}()

	// handle shutdown
	select {
	case <-ctx.Done():
	case <-s.ctx.Done():
		s.log.Info("Service.Run: interrupt syscall received")
		cancel()
	}

	s.log.Info("Service.Run: cleaning up…")
	s.listener.Close()
	s.log.Info("Service.Run: cleanup done, waiting workers to stop…")
	wg.Wait()
	s.log.Info("Service.Run: shutdown")
}

func (s *Service) handleConn(conn net.Conn) {
	defer conn.Close()

	if s.cfg.connReadTTL > 0 {
		err := conn.SetReadDeadline(time.Now().Add(s.cfg.connReadTTL))
		if err != nil {
			err = fmt.Errorf(
				"Service.handleConn: can't set read deadline: %w", err)
			s.log.Error(err)
			return
		}
	}
	if s.cfg.connWriteTTL > 0 {
		err := conn.SetWriteDeadline(time.Now().Add(s.cfg.connWriteTTL))
		if err != nil {
			err = fmt.Errorf(
				"Service.handleConn: can't set write deadline: %w", err)
			s.log.Error(err)
			return
		}
	}

	s.log.Debugf(
		"Service.handleConn: accept client conn %q",
		conn.RemoteAddr().String())
	defer func() {
		s.log.Debugf(
			"Service.handleConn: close client conn %q",
			conn.RemoteAddr().String())
	}()

	for {
		dec := gob.NewDecoder(conn)
		message := proto.Message{}
		err := dec.Decode(&message)
		if err != nil {
			err = fmt.Errorf(
				"Service.handleConn: decode message error: %w", err)
			s.log.Error(err)
			return
		}

		switch message.MessageType {
		case proto.RequestServiceType:
			s.log.Debug("Service.handleConn: got request service")
			err = s.handleRequestService(conn)
			if err != nil {
				err = fmt.Errorf("Service.handleConn: %w", err)
				s.log.Error(err)
				err = s.handleError(conn, "try again later")
				if err != nil {
					err = fmt.Errorf("Service.handleConn: %w", err)
					s.log.Error(err)
				}
				return
			}
		case proto.ResponseChallengeType:
			s.log.Debug("Service.handleConn: got response challenge")
			err = s.handleResponseChallenge(conn, message.Data)
			if err != nil {
				err = fmt.Errorf("Service.handleConn: %w", err)
				s.log.Error(err)
				err = s.handleError(conn, "try again later")
				if err != nil {
					err = fmt.Errorf("Service.handleConn: %w", err)
					s.log.Error(err)
				}
				return
			}
		default:
			s.log.Debugf(
				"Service.handleConn: got unknown message type: %d",
				message.MessageType)
			err = s.handleError(
				conn, fmt.Sprintf("unknown message type: %d", message.MessageType))
			if err != nil {
				err = fmt.Errorf("Service.handleConn: %w", err)
				s.log.Error(err)
			}
			return
		}
	}
}

func (s *Service) adjustDifficultyByRPS() {
	s.mux.Lock()
	rps := s.rps
	difficulty := s.difficulty
	if s.rps < s.cfg.rpsTarget && s.difficulty > 1 {
		s.difficulty -= 1
	} else if s.rps > s.cfg.rpsTarget && s.difficulty < pow.MaxDifficulty {
		s.difficulty += 1
	}
	s.rps = 0
	s.mux.Unlock()

	s.log.Debugf("difficulty: %d, rps: %d", difficulty, rps)
}

func (s *Service) adjustDifficulty(expires time.Time) {
	from := expires.Add(-s.cfg.challengeTTL)
	if time.Now().Sub(from) < s.cfg.challengeTTL {
		s.mux.Lock()
		if s.difficulty < pow.MaxDifficulty {
			s.difficulty += 1
		}
		s.mux.Unlock()
	} else {
		s.mux.Lock()
		if s.difficulty > 1 {
			s.difficulty -= 1
		}
		s.mux.Unlock()
	}
}

// logger desribes interface of log object.
type logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warning(...interface{})
	Warningf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
}
