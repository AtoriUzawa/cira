package cira

import (
	"errors"
	"sync"

	"github.com/AtoriUzawa/cira/internal/idgen"
	"github.com/AtoriUzawa/cira/internal/protocol"
	"github.com/gorilla/websocket"
)

type hub struct {
	peers map[string]*peer

	router *Router

	idGenerator idgen.IDGenerator
	codec       protocol.Codec

	OnPeerClose func(string)

	mu sync.Mutex
}

// ErrConnNotFound is returned when a connection is not found in the hub.
var ErrConnNotFound = errors.New("ws/hub: connection not found")

func newHub(router *Router, idGenerator idgen.IDGenerator, codec protocol.Codec) *hub {
	return &hub{
		peers:       make(map[string]*peer, 0),
		router:      router,
		idGenerator: idGenerator,
		codec:       codec,
	}
}

func (h *hub) Register(wc *websocket.Conn) *peer {
	peer := newPeer(wc, h.idGenerator.Next())
	peer.Dispatcher = h
	peer.Codec = h.codec
	peer.OnClose(func() { h.Unregister(peer.ID()) })

	h.mu.Lock()
	h.peers[peer.ID()] = peer
	h.mu.Unlock()

	return peer
}

func (h *hub) Unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.peers[id]
	if ok {
		delete(h.peers, id)
	}

	// if ok {
	// 	peer.Close()
	// }
}

func (h *hub) Dispatch(c *peer, m *Message) bool {
	ctx := newContext(c, h.idGenerator, h.codec)
	ctx.Message = m
	ctx.Conn = newConn(c, h)
	return h.router.dispatch(ctx)
}

func (h *hub) peer(id string) (*peer, bool) {
	peer, ok := h.peers[id]
	if !ok {
		return nil, false
	}

	return peer, true
}
