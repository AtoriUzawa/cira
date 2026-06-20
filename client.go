// Package cira provides a lightweight, event-driven WebSocket framework.
package cira

import (
	"errors"
	"fmt"

	"github.com/AtoriUzawa/cira/internal/protocol"
	"github.com/AtoriUzawa/cira/internal/runtime"
	"github.com/AtoriUzawa/cira/internal/transport"
	"github.com/gorilla/websocket"
)

type client struct {
	id string

	conn *transport.Conn

	Runtime *runtime.Runtime

	Codec protocol.Codec

	Dispatcher Dispatcher
}

// Dispatcher routes incoming messages to the appropriate handler.
type Dispatcher interface {
	Dispatch(*client, *Message) bool
}

// New creates a client.
func newClient(ws *websocket.Conn, id string) *client {
	conn := transport.New(ws, id)
	c := &client{
		id:      id,
		conn:    conn,
		Runtime: runtime.NewRuntime(),
	}

	conn.OnMessage = c.onMessage

	return c
}

// ErrInvalidMessageType is the error message prefix for unrecognized message types.
var ErrInvalidMessageType = "ws/client: invalid message type"

// ID returns the client id.
func (c *client) ID() string {
	return c.id
}

// Start starts the client.
func (c *client) Start() {
	c.conn.Start()
}

// Send sends a message.
func (c *client) Send(data []byte) {
	c.conn.Send(data)
}

func (c *client) Close() {
	c.conn.Close()
}

func (c *client) OnClose(fn func()) {
	c.conn.OnClose = append(c.conn.OnClose, fn)
}

func (c *client) onMessage(data []byte) error {
	var msg Message

	err := c.Codec.Decode(data, &msg)
	if err != nil {
		return fmt.Errorf(ErrInvalidMessageType+"; %w", err)
	}

	if msg.Type == TypeResp &&
		c.Runtime.Resolve(msg.ReplyTo, &runtime.CallResult{Data: msg.Data}) {
		return nil
	}

	if ok := c.Dispatcher.Dispatch(c, &msg); !ok {
		return errors.New(ErrInvalidMessageType + "; can not dispatch")
	}

	return nil
}
