package cira

import (
	"errors"
	"net/http"

	"github.com/AtoriUzawa/cira/internal/idgen"
	"github.com/AtoriUzawa/cira/internal/protocol"
	"github.com/gorilla/websocket"
)

// Engine is the core WebSocket server that manages connections, routing, and lifecycle.
type Engine struct {
	*RouterGroup

	hub *hub

	upgrader websocket.Upgrader

	idGenerator idgen.IDGenerator

	codec protocol.Codec

	OnConnect func(*Conn)
}

// Option configures an Engine with optional settings.
type Option func(*Engine)

// ErrNotFoundConn is returned when no connection matches the requested ID.
var ErrNotFoundConn = errors.New("ws/engine: not found this conn")

// New creates a new Engine with the given options.
func New(opts ...Option) *Engine {
	router := newRouter()

	engine := &Engine{
		idGenerator: &idgen.UUIDGenerator{},

		codec: &protocol.JSONCodec{},

		upgrader: websocket.Upgrader{
			CheckOrigin: func(
				r *http.Request,
			) bool {
				return true
			},
		},
	}

	engine.RouterGroup = &RouterGroup{
		router: router,
	}

	for _, opt := range opts {
		opt(engine)
	}

	engine.hub = newHub(
		router,
		engine.idGenerator,
		engine.codec,
	)

	return engine
}

// HandleWS upgrades an HTTP connection to WebSocket and starts handling it.
func (e *Engine) HandleWS(
	w http.ResponseWriter,
	r *http.Request,
) {
	conn, err := e.upgrader.Upgrade(
		w,
		r,
		nil,
	)
	if err != nil {
		return
	}

	peer := e.hub.Register(conn)

	go peer.Start()

	if e.OnConnect != nil {
		e.OnConnect(
			newConn(
				peer,
				e.hub,
			),
		)
	}
}

// Run starts the HTTP server on the given address and handles WebSocket connections at /ws.
func (e *Engine) Run(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/ws",
		e.HandleWS,
	)

	return http.ListenAndServe(
		addr,
		mux,
	)
}

func (e *Engine) Dial(url string) (*Conn, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	peer := e.hub.Register(ws)
	conn := newConn(peer, e.hub)

	go peer.Start()

	return conn, nil
}

// Conn returns the connection with the given ID, or ErrNotFoundConn if it does not exist.
func (e *Engine) Conn(
	id string,
) (*Conn, error) {
	p, ok := e.hub.peer(id)

	if !ok {
		return nil, ErrNotFoundConn
	}

	return newConn(
		p,
		e.hub,
	), nil
}

// WithIDGenerator sets a custom ID generator for the Engine.
func WithIDGenerator(g idgen.IDGenerator) Option {
	return func(s *Engine) {
		s.idGenerator = g
	}
}

// WithCodec sets a custom codec for message serialization.
func WithCodec(c protocol.Codec) Option {
	return func(s *Engine) {
		s.codec = c
	}
}

// WithUpgrader sets a custom WebSocket upgrader for the Engine.
func WithUpgrader(u websocket.Upgrader) Option {
	return func(s *Engine) {
		s.upgrader = u
	}
}
