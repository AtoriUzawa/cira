package cira

import (
	"errors"
	"sync"

	"github.com/AtoriUzawa/cira/internal/idgen"
	"github.com/AtoriUzawa/cira/internal/protocol"
	"github.com/gorilla/websocket"
)

type hub struct {
	clients map[string]*client

	router *Router

	idGenerator idgen.IDGenerator
	codec       protocol.Codec

	OnClientClose func(string)

	mu sync.Mutex
}

// ErrConnNotFound is returned when a connection is not found in the hub.
var ErrConnNotFound = errors.New("ws/hub: connection not found")

func newHub(router *Router, idGenerator idgen.IDGenerator, codec protocol.Codec) *hub {
	return &hub{
		clients:     make(map[string]*client, 0),
		router:      router,
		idGenerator: idGenerator,
		codec:       codec,
	}
}

func (h *hub) Register(wc *websocket.Conn) *client {
	client := newClient(wc, h.idGenerator.Next())
	client.Dispatcher = h
	client.Codec = h.codec
	client.OnClose(func() { h.Unregister(client.ID()) })

	h.mu.Lock()
	h.clients[client.ID()] = client
	h.mu.Unlock()

	return client
}

func (h *hub) Unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.clients[id]
	if ok {
		delete(h.clients, id)
	}

	// if ok {
	// 	client.Close()
	// }
}

func (h *hub) Dispatch(c *client, m *Message) bool {
	ctx := newContext(c, h.idGenerator, h.codec)
	ctx.Message = m
	ctx.Conn = newConn(c, h)
	return h.router.dispatch(ctx)
}

func (h *hub) client(id string) (*client, bool) {
	client, ok := h.clients[id]
	if !ok {
		return nil, false
	}

	return client, true
}
