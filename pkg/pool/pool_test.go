package pool

import (
	"context"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := New(ctx, time.Second, time.Second)
	_ = c.Set("key")

	if len(c.data) != 1 {
		t.Errorf("expected pool data len to be 1 but got %d", len(c.data))
	}

	// check we can get the key.
	ok := c.Get("key")
	if !ok {
		t.Error("get should return 'true' value")
	}

	// check the key does not exist.
	ok = c.Get("key")
	if ok {
		t.Errorf("get should return 'false' value")
	}
}
