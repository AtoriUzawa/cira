// Package cira provides a lightweight, event-driven WebSocket framework.
package cira

import (
	"errors"
	"fmt"
	"io"

	"github.com/AtoriUzawa/cira/internal/protocol"
	"github.com/AtoriUzawa/cira/internal/runtime"
	"github.com/AtoriUzawa/cira/internal/transport"
	"github.com/gorilla/websocket"
)

type peer struct {
	id string

	conn *transport.Conn

	Runtime *runtime.Runtime

	Codec protocol.Codec

	Dispatcher Dispatcher
}

// Dispatcher routes incoming messages to the appropriate handler.
type Dispatcher interface {
	Dispatch(*peer, *Message) bool
}

// New creates a peer.
func newPeer(ws *websocket.Conn, id string) *peer {
	conn := transport.New(ws, id)
	c := &peer{
		id:      id,
		conn:    conn,
		Runtime: runtime.NewRuntime(),
	}

	conn.OnMessage = c.onMessage

	return c
}

// ErrInvalidMessageType is the error message prefix for unrecognized message types.
var ErrInvalidMessageType = "ws/peer: invalid message type"

// ID returns the peer id.
func (c *peer) ID() string {
	return c.id
}

// Start starts the peer.
func (c *peer) Start() {
	c.conn.Start()
}

// Send sends a message.
func (c *peer) Send(data []byte) {
	c.conn.Send(data)
}

func (c *peer) Close() {
	c.conn.Close()
}

func (c *peer) OnClose(fn func()) {
	c.conn.OnClose = append(c.conn.OnClose, fn)
}

func (c *peer) onMessage(data []byte) error {
	var msg Message

	err := c.Codec.Decode(data, &msg)
	if err != nil {
		return fmt.Errorf(ErrInvalidMessageType+"; %w", err)
	}

	if msg.Type == TypeResp &&
		c.Runtime.Resolve(msg.ReplyTo, &runtime.Delivery{Data: msg.Data}) {
		c.Runtime.Unregister(msg.ReplyTo)
		return nil
	}

	if msg.Type == TypeStream {
		c.Runtime.Resolve(msg.ReplyTo, &runtime.Delivery{Data: msg.Data})
		return nil
	}

	if msg.Type == TypeStreamClose {
		c.Runtime.Resolve(msg.ReplyTo, &runtime.Delivery{Err: io.EOF})
		c.Runtime.Unregister(msg.ReplyTo)
		return nil
	}

	if ok := c.Dispatcher.Dispatch(c, &msg); !ok {
		return errors.New(ErrInvalidMessageType + "; can not dispatch")
	}

	return nil
}
