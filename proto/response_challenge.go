package proto

// ResponseChallenge struct holds nonce and hash values for signed requested challenge data.
type ResponseChallenge struct {
	Nonce    uint64
	Hash     []byte
	Envelope *Envelope
}
