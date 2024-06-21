package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math"
	"math/bits"
	"time"
)

const (
	MinDifficulty = 1
	MaxDifficulty = 255

	nonceBytesLen = 8
	proofBytesLen = 32
)

// Fulfil returns valid hash and nonce value
// for specified challenge of specified difficulty.
func Fulfil(
	challenge []byte, difficulty int, expires time.Time,
) ([]byte, uint64) {
	var (
		nonce        uint64
		hasher       = sha256.New()
		leadingZeros int
		resultBytes  []byte
	)

	for {
		// check if time to solve is out
		if time.Now().After(expires) {
			return nil, 0
		}
		// overflow overflow
		if nonce == math.MaxUint64 {
			return nil, 0
		}

		resultBytes, leadingZeros = calc(hasher, challenge, nonce)
		if leadingZeros >= difficulty {
			return resultBytes, nonce
		}

		hasher.Reset()
		nonce++
	}
}

// Check returns true if proof is valid.
func Check(
	challenge []byte, difficulty int,
	proof []byte, nonce uint64,
) bool {
	// validate input parameters
	if difficulty > MaxDifficulty {
		return false
	}
	if len(proof) != proofBytesLen {
		return false
	}

	hasher := sha256.New()
	resultBytes, leadingZeros := calc(hasher, challenge, nonce)
	if leadingZeros < difficulty {
		return false
	}

	return bytes.Compare(proof, resultBytes) == 0
}

func calc(hasher hash.Hash, challenge []byte, nonce uint64) ([]byte, int) {
	var (
		nonceBytes  = make([]byte, nonceBytesLen)
		resultBytes = make([]byte, proofBytesLen)
	)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	hasher.Write(challenge)
	hasher.Write(nonceBytes)
	hasher.Sum(resultBytes[:0])
	v := binary.BigEndian.Uint64(resultBytes)
	return resultBytes, bits.LeadingZeros64(v)
}
