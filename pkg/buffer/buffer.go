package buffer

import (
	"crypto/rand"
	"sync"
)

// defaultChallengeLen controls len of each pre-calculated challenge value.
const defaultChallengeLen = 8

// defaultBufferSize controls amount of pre-calculated challenges in buffer.
const defaultBufferSize = 100

type rec struct {
	prev  *rec
	next  *rec
	value []byte
}

// Buffer is the thread-safe linked-list-based pool of pre-calculated
// challenges to be used in challenge requests.
type Buffer struct {
	log          logger
	head         *rec
	tail         *rec
	challengeLen int
	bufferSize   int
	mux          sync.RWMutex
}

// New creates Buffer instance and init it with pre-calculated challenges
// in quantity specified by bufferSize.
func New(log logger, challengeLen, bufferSize int) (*Buffer, error) {
	if challengeLen == 0 {
		challengeLen = defaultChallengeLen
	}
	if bufferSize == 0 {
		bufferSize = defaultBufferSize
	}

	buf := &Buffer{
		log:          log,
		challengeLen: challengeLen,
		bufferSize:   bufferSize,
	}

	for i := 0; i < bufferSize; i++ {
		err := buf.gen()
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
}

// gen generates new challenge specified by len.
func (buf *Buffer) gen() error {
	b, err := rnd(buf.challengeLen)
	if err != nil {
		return err
	}

	buf.mux.Lock()
	defer buf.mux.Unlock()

	buf.bufferSize += 1
	rec := &rec{value: b}

	if buf.head == nil && buf.tail == nil {
		buf.head = rec
		buf.tail = rec
		return nil
	}

	rec.prev = buf.tail
	buf.tail.next = rec
	buf.tail = rec

	return nil
}

// Pop returns pre-calculated challenge, removes it from the buffer
// and spawns background task to generate new one.
func (buf *Buffer) Pop() []byte {
	buf.mux.Lock()
	defer buf.mux.Unlock()

	if buf.head == nil {
		return nil
	}

	buf.bufferSize -= 1

	head := buf.head
	buf.head = buf.head.next
	if buf.head != nil {
		buf.head.prev = nil
	}

	go func() {
		err := buf.gen()
		if err != nil {
			buf.log.Error(err)
		}
	}()

	return head.value
}

// rnd returns random bytes slice specified with length.
func rnd(l int) ([]byte, error) {
	b := make([]byte, l)

	cnt := 0
	for l > cnt {
		n, err := rand.Read(b[cnt:])
		if err != nil {
			return nil, err
		}
		cnt += n
	}

	return b, nil
}
