package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/bits"
)

// Fulfil returns valid hash and nonce value
// for specified challenge of specified difficulty.
func Fulfil(
	log logger, challenge []byte, difficulty int,
) ([]byte, uint64) {
	var (
		nonce, v     uint64 = 0, 0
		nonceBytes          = make([]byte, 8)
		resultBytes         = make([]byte, 32)
		hasher              = sha256.New()
		leadingZeros int    = 0
	)

	log.Debug("begin to fulfil challenge")
	for {
		binary.BigEndian.PutUint64(nonceBytes, nonce)
		hasher.Write(challenge)
		hasher.Write(nonceBytes)
		hasher.Sum(resultBytes[:0])

		v = binary.BigEndian.Uint64(resultBytes)
		leadingZeros = bits.LeadingZeros64(v)
		if leadingZeros >= difficulty {
			log.Debugf("challenge fulfiled with nonce %d", nonce)
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
	if difficulty > 255 {
		return false
	}
	if len(proof) != 32 {
		return false
	}

	var (
		v            uint64 = 0
		nonceBytes          = make([]byte, 8)
		resultBytes         = make([]byte, 32)
		hasher              = sha256.New()
		leadingZeros int    = 0
	)

	binary.BigEndian.PutUint64(nonceBytes, nonce)
	hasher.Write(challenge)
	hasher.Write(nonceBytes)
	hasher.Sum(resultBytes[:0])

	v = binary.BigEndian.Uint64(resultBytes)
	leadingZeros = bits.LeadingZeros(uint(v))
	if leadingZeros < difficulty {
		return false
	}

	return bytes.Compare(proof, resultBytes) == 0
}
