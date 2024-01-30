package server

import "sync"

type quotes struct {
	data []string
	idx  int
	mux  sync.Mutex
}

func newQuotes() *quotes {
	data := []string{
		"Don't be a Jack of all trades and master of none.",
		"Be sticky to your research goal.",
	}

	q := &quotes{data: data}

	return q
}

func (q *quotes) get() string {
	q.mux.Lock()
	s := q.data[q.idx]
	q.idx++
	if q.idx == len(q.data) {
		q.idx = 0
	}
	q.mux.Unlock()

	return s
}
