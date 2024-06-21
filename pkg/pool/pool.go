package pool

import (
	"context"
	"sync"
	"time"
)

// defaultCleanupInterval controls how often active challenge registry pool
// will purge obsolete values.
const defaultCleanupInterval = time.Second

type ttlRec struct {
	// linked list to maintain order
	prev      string
	next      string
	timestamp time.Time
}

// Pool is the thread-safe map-based pool with TTL
// record invalidation support.
// Pool uses double linked list to maintain FIFO order
// of inserted values.
type Pool struct {
	data map[string]ttlRec
	mux  sync.RWMutex
	ttl  time.Duration
	tail string
	head string
}

// New creates Pool instance
// and spawns background cleanup goroutine,
// that periodically removes outdated records.
// cleanup goroutine will run cleanup once in cleanupInterval until ctx is canceled.
// each record in the pool is valid for ttl duration since it was set.
func New(
	ctx context.Context,
	ttl time.Duration,
	cleanupInterval time.Duration,
) *Pool {
	if cleanupInterval == 0 {
		cleanupInterval = defaultCleanupInterval
	}
	c := Pool{
		data: make(map[string]ttlRec),
		ttl:  ttl,
	}

	go func(ctx context.Context) {
		t := time.NewTicker(cleanupInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_ = c.cleanup()
			}
		}
	}(ctx)

	return &c
}

// Set adds the key to registry and returns expire timestamp.
func (c *Pool) Set(key string) time.Time {
	c.mux.Lock()
	defer c.mux.Unlock()

	val := ttlRec{
		prev:      c.tail,
		timestamp: time.Now(),
	}

	if c.head == "" {
		c.head = key
		c.tail = key
		val.prev = ""
		c.data[key] = val

		return val.timestamp.Add(c.ttl)
	}

	// If the record for this key already exists
	// and is somewhere in the middle of the list
	// removing it before adding to the tail.
	if rec, ok := c.data[key]; ok && key != c.tail {
		prev := c.data[rec.prev]
		next := c.data[rec.next]
		prev.next = rec.next
		next.prev = rec.prev
		c.data[rec.prev] = prev
		c.data[rec.next] = next
	}

	tailval := c.data[c.tail]
	tailval.next = key
	c.data[c.tail] = tailval
	c.tail = key
	c.data[key] = val

	return val.timestamp.Add(c.ttl)
}

// Get returns false if key is not found in the pool or record is outdated,
// otherwise it returns true and removes the key from pool.
func (c *Pool) Get(key string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	rec, ok := c.data[key]
	if !ok {
		return ok
	}

	if time.Now().Sub(rec.timestamp) >= c.ttl {
		return false
	}

	delete(c.data, key)

	if key == c.head {
		c.head = rec.next
	}

	if key == c.tail {
		c.tail = rec.prev
	}

	if rec.prev != "" {
		prev := c.data[rec.prev]
		prev.next = rec.next
		c.data[rec.prev] = prev
	}

	if rec.next != "" {
		next := c.data[rec.next]
		next.prev = rec.prev
		c.data[rec.next] = next
	}

	return true
}

// cleanup removes outdated records.
func (c *Pool) cleanup() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	key := c.head
	for {
		rec, ok := c.data[key]
		if !ok {
			break
		}

		if time.Now().Sub(rec.timestamp) < c.ttl {
			break
		}

		c.head = rec.next
		delete(c.data, key)

		if key == c.tail {
			c.tail = ""
			return nil
		}

		next, ok := c.data[rec.next]
		if ok {
			next.prev = ""
			c.data[rec.next] = next
		}
		key = rec.next
	}

	return nil
}

// Len returns the number of records in the registry.
func (c *Pool) Len() int {
	c.mux.RLock()
	defer c.mux.RUnlock()

	return len(c.data)
}
