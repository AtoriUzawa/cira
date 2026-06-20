package cira

type (
	// HandlerFunc processes a WebSocket message context.
	HandlerFunc func(*Context)
	// Middleware wraps a HandlerFunc to add cross-cutting behavior.
	Middleware func(HandlerFunc) HandlerFunc
)
