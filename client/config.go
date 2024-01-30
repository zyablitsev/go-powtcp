package client

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultLogLevel     = "debug"
	defaultConnReadTTL  = "1s"
	defaultConnWriteTTL = "1s"
)

type config struct {
	logLevel     string
	serverAddr   string
	connReadTTL  time.Duration
	connWriteTTL time.Duration
}

func loadConfig() (config, error) {
	logLevelEnv := os.Getenv("CLIENT_LOG_LEVEL")
	if logLevelEnv == "" {
		logLevelEnv = defaultLogLevel
	}

	serverAddrEnv := os.Getenv("CLIENT_SERVER_ADDR")
	if serverAddrEnv == "" {
		err := fmt.Errorf("bad CLIENT_SERVER_ADDR env value %q", serverAddrEnv)
		return config{}, err
	}

	connReadTTLEnv := os.Getenv("CLIENT_CONNREAD_TTL")
	if connReadTTLEnv == "" {
		connReadTTLEnv = defaultConnReadTTL
	}
	connReadTTL, err := time.ParseDuration(connReadTTLEnv)
	if err != nil {
		err = fmt.Errorf(
			"bad CLIENT_CONNREAD_TTL env value %q: %v",
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
			"bad CLIENT_CONNWRITE_TTL env value %q: %v",
			connWriteTTLEnv, err)
		return config{}, err
	}

	cfg := config{
		logLevel:     logLevelEnv,
		serverAddr:   serverAddrEnv,
		connReadTTL:  connReadTTL,
		connWriteTTL: connWriteTTL,
	}

	return cfg, nil
}
