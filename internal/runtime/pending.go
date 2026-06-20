// Package runtime
package runtime

import (
	"sync"
)

// Runtime manages pending RPC call registrations and their results.
type Runtime struct {
	pendings map[string]chan *CallResult
	mu       sync.Mutex
}

// CallResult holds the result of a pending RPC call, containing the response data or an error.
type CallResult struct {
	Data []byte
	Err  error
}

// NewRuntime creates and returns a new Runtime instance with an empty pending map.
func NewRuntime() *Runtime {
	return &Runtime{
		pendings: make(map[string]chan *CallResult, 0),
	}
}

// Register registers a pending call with the given ID and returns a channel to receive the result.
func (r *Runtime) Register(id string) <-chan *CallResult {
	ch := make(chan *CallResult, 1)

	r.mu.Lock()
	r.pendings[id] = ch
	r.mu.Unlock()

	return ch
}

// Unregister removes a pending call registration for the given ID.
func (r *Runtime) Unregister(id string) {
	r.mu.Lock()
	delete(r.pendings, id)
	r.mu.Unlock()
}

// Resolve delivers a result to the pending call identified by id. It returns true if the call was found and resolved.
func (r *Runtime) Resolve(id string, res *CallResult) bool {
	r.mu.Lock()
	ch, ok := r.pendings[id]
	if ok {
		delete(r.pendings, id)
	}
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
