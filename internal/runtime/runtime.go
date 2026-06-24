// Package runtime
package runtime

import (
	"sync"
)

// Runtime route incoming message deliveries to registered consumers.
type Runtime struct {
	delivery map[string]chan *Delivery
	mu       sync.Mutex
}

// Delivery represents a runtime-delivered message payload or terminal error.
type Delivery struct {
	Data []byte
	Err  error
}

// NewRuntime creates and returns a new Runtime instance with an empty pending map.
func NewRuntime() *Runtime {
	return &Runtime{
		delivery: make(map[string]chan *Delivery, 0),
	}
}

// Register creates a delivery channel for the specified identifier.
func (r *Runtime) Register(id string) <-chan *Delivery {
	ch := make(chan *Delivery, 1)

	r.mu.Lock()
	r.delivery[id] = ch
	r.mu.Unlock()

	return ch
}

// Unregister removes the delivery registration associated with the identifier.
func (r *Runtime) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.delivery, id)
}

// Resolve delivers a message to the registered consumer identified by id.
// It returns true if a matching registration was found.
func (r *Runtime) Resolve(id string, res *Delivery) bool {
	r.mu.Lock()
	ch, ok := r.delivery[id]
	r.mu.Unlock()

	if !ok {
		return false
	}

	select {
	case ch <- res:
	default:
	}

	return true
}
