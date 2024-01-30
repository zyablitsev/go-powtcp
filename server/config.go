package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	defaultLogLevel                            = "debug"
	defaultSecret                              = "secret"
	defaultPort                         uint16 = 9999
	defaultRPSTarget                    int    = 1000
	defaultDifficulty                   int    = 5
	defaultChallengeLen                 int    = 8
	defaultChallengeBufferSize          int    = 100
	defaultChallengeTTL                        = "1s"
	defaultChallengePoolCleanupInterval        = "1s"
	defaultConnReadTTL                         = "1s"
	defaultConnWriteTTL                        = "1s"
)

var (
	defaultIP = net.IPv4(0, 0, 0, 0)
)

type config struct {
	logLevel                     string
	secret                       string
	ip                           net.IP
	port                         uint16
	rpsTarget                    int
	difficulty                   int
	challengeLen                 int
	challengeBufferSize          int
	challengeTTL                 time.Duration
	challengePoolCleanupInterval time.Duration
	connReadTTL                  time.Duration
	connWriteTTL                 time.Duration
}

func loadConfig() (config, error) {
	logLevelEnv := os.Getenv("SERVER_LOG_LEVEL")
	if logLevelEnv == "" {
		logLevelEnv = defaultLogLevel
	}

	secretEnv := os.Getenv("SERVER_SECRET")
	if secretEnv == "" {
		secretEnv = defaultSecret
	}

	ipEnv := os.Getenv("SERVER_IP")
	ip := net.ParseIP(ipEnv).To4()
	if ip == nil && ipEnv != "" {
		return config{}, fmt.Errorf("bad SERVER_IP env value %q", ipEnv)
	}
	if ip == nil {
		ip = defaultIP
	}

	portEnv := os.Getenv("SERVER_PORT")
	var port uint16
	if portEnv == "" {
		port = defaultPort
	} else {
		v, err := strconv.ParseUint(portEnv, 10, 16)
		if err != nil {
			err = fmt.Errorf("bad SERVER_PORT env value %q: %v", portEnv, err)
			return config{}, err
		}
		port = uint16(v)
	}

	rpsTargetEnv := os.Getenv("SERVER_RPS_TARGET")
	var rpsTarget int
	if rpsTargetEnv == "" {
		rpsTarget = defaultRPSTarget
	} else {
		v, err := strconv.ParseInt(rpsTargetEnv, 10, 64)
		if err != nil {
			err = fmt.Errorf(
				"bad SERVER_RPS_TARGET env value %q: %v",
				rpsTargetEnv, err)
			return config{}, err
		}
		rpsTarget = int(v)
		if rpsTarget < 1 {
			err = fmt.Errorf(
				"bad SERVER_RPS_TARGET env value %q: should be gt 0",
				rpsTargetEnv)
			return config{}, err
		}
	}

	difficultyEnv := os.Getenv("SERVER_DIFFICULTY")
	var difficulty int
	if difficultyEnv == "" {
		difficulty = defaultDifficulty
	} else {
		v, err := strconv.ParseInt(difficultyEnv, 10, 8)
		if err != nil {
			err = fmt.Errorf(
				"bad SERVER_DIFFICULTY env value %q: %v",
				difficultyEnv, err)
			return config{}, err
		}
		difficulty = int(v)
		if difficulty < 1 {
			err = fmt.Errorf(
				"bad SERVER_DIFFICULTY env value %q: should be gt 0",
				difficultyEnv)
			return config{}, err
		}
	}

	challengeLenEnv := os.Getenv("SERVER_CHALLENGE_LEN")
	var challengeLen int
	if challengeLenEnv == "" {
		challengeLen = defaultChallengeLen
	} else {
		v, err := strconv.ParseInt(challengeLenEnv, 10, 64)
		if err != nil {
			err = fmt.Errorf(
				"bad SERVER_CHALLENGE_LEN env value %q: %v",
				challengeLenEnv, err)
			return config{}, err
		}
		challengeLen = int(v)
		if challengeLen < 1 {
			err = fmt.Errorf(
				"bad SERVER_CHALLENGE_LEN env value %q: should be gt 0",
				challengeLenEnv)
			return config{}, err
		}
	}

	challengeBufferSizeEnv := os.Getenv("SERVER_CHALLENGE_BUFFER_SIZE")
	var challengeBufferSize int
	if challengeBufferSizeEnv == "" {
		challengeBufferSize = defaultChallengeBufferSize
	} else {
		v, err := strconv.ParseInt(
			challengeBufferSizeEnv, 10, 64)
		if err != nil {
			err = fmt.Errorf(
				"bad SERVER_CHALLENGE_BUFFER_SIZE env value %q: %v",
				challengeBufferSizeEnv, err)
			return config{}, err
		}
		challengeBufferSize = int(v)
		if challengeBufferSize < 1 {
			err = fmt.Errorf(
				"bad SERVER_CHALLENGE_BUFFER_SIZE env value %q: should be gt 0",
				challengeBufferSizeEnv)
			return config{}, err
		}
	}

	challengeTTLEnv := os.Getenv("SERVER_CHALLENGE_TTL")
	if challengeTTLEnv == "" {
		challengeTTLEnv = defaultLogLevel
	}
	challengeTTL, err := time.ParseDuration(challengeTTLEnv)
	if err != nil {
		err = fmt.Errorf(
			"bad SERVER_CHALLENGE_TTL env value %q: %v",
			challengeTTLEnv, err)
		return config{}, err
	}

	challengePoolCleanupIntervalEnv := os.Getenv(
		"SERVER_CHALLENGE_POOL_CLEANUP_INTERVAL")
	if challengePoolCleanupIntervalEnv == "" {
		challengePoolCleanupIntervalEnv = defaultChallengePoolCleanupInterval
	}
	challengePoolCleanupInterval, err := time.ParseDuration(
		challengePoolCleanupIntervalEnv)
	if err != nil {
		err = fmt.Errorf(
			"bad SERVER_CHALLENGE_POOL_CLEANUP_INTERVAL env value %q: %v",
			challengePoolCleanupIntervalEnv, err)
		return config{}, err
	}

	connReadTTLEnv := os.Getenv("SERVER_CONNREAD_TTL")
	if connReadTTLEnv == "" {
		connReadTTLEnv = defaultConnReadTTL
	}
	connReadTTL, err := time.ParseDuration(connReadTTLEnv)
	if err != nil {
		err = fmt.Errorf(
			"bad SERVER_CONNREAD_TTL env value %q: %v",
			connReadTTLEnv, err)
		return config{}, err
	}

	connWriteTTLEnv := os.Getenv("SERVER_CONNWRITE_TTL")
	if connWriteTTLEnv == "" {
		connWriteTTLEnv = defaultConnWriteTTL
	}
	connWriteTTL, err := time.ParseDuration(connWriteTTLEnv)
	if err != nil {
		err = fmt.Errorf(
			"bad SERVER_CONNWRITE_TTL env value %q: %v",
			connWriteTTLEnv, err)
		return config{}, err
	}

	cfg := config{
		logLevel:                     logLevelEnv,
		secret:                       secretEnv,
		ip:                           ip,
		port:                         uint16(port),
		rpsTarget:                    rpsTarget,
		difficulty:                   difficulty,
		challengeLen:                 challengeLen,
		challengeBufferSize:          challengeBufferSize,
		challengeTTL:                 challengeTTL,
		challengePoolCleanupInterval: challengePoolCleanupInterval,
		connReadTTL:                  connReadTTL,
		connWriteTTL:                 connWriteTTL,
	}

	return cfg, nil
}
