// Package transport provides websocket connection encapsulation.
package transport

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/AtoriUzawa/cira/internal/runtime"
	"github.com/gorilla/websocket"
)

const (
	// WriteChCapacity is the capacity of the write channel buffer.
	WriteChCapacity = 10

	// PingInterval is the interval at which WebSocket ping frames are sent.
	PingInterval = 10 * time.Second

	// PongWait is the maximum duration to wait for a pong response before closing.
	PongWait = 30 * time.Second

	// WriteWait is the maximum duration allowed for a write operation to complete.
	WriteWait = 10 * time.Second
)

var (
	// ErrWriteChanFull is returned when the write channel has reached its capacity.
	ErrWriteChanFull = errors.New("ws/conn: write chan full")

	// ErrInvalidFrameType is returned when a received WebSocket frame is not a text message.
	ErrInvalidFrameType = errors.New("ws/conn: invalid frame type")
)

// Conn encapsulates the underlying websocket connection.
//
// Responsibilities:
//  1. Manage websocket lifecycle.
//  2. Manage read/write pump.
//  3. Manage ping/pong.
//  4. Provide []byte send/receive.
type Conn struct {
	ws *websocket.Conn

	writeCh chan []byte

	ID        string
	OnMessage func([]byte) error
	OnClose   []func()

	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// New creates a websocket conn.
func New(ws *websocket.Conn, id string) *Conn {
	ctx, cancel := context.WithCancel(context.Background())

	return &Conn{
		ws:      ws,
		writeCh: make(chan []byte, WriteChCapacity),
		ID:      id,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the read/write loop.
func (c *Conn) Start() {
	go c.readPump()
	go c.writePump()
}

// Send asynchronously sends a message.
func (c *Conn) Send(data []byte) {
	select {
	case c.writeCh <- data:

	case <-c.ctx.Done():
		return

	default:
		c.handleErr(runtime.OpSendChan, data, ErrWriteChanFull)
	}
}

// Close closes the connection.
func (c *Conn) Close() {
	c.once.Do(func() {
		c.cancel()

		_ = c.ws.Close()

		for _, fn := range c.OnClose {
			fn()
		}
	})
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(PingInterval)

	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case data := <-c.writeCh:
			if err := c.ws.SetWriteDeadline(
				time.Now().Add(WriteWait),
			); err != nil {
				c.handleErr(runtime.OpSetWriteDeadline, data, err)
				return
			}

			if err := c.write(data); err != nil {
				c.handleErr(runtime.OpWrite, data, err)
				return
			}

		case <-ticker.C:
			if err := c.ws.SetWriteDeadline(
				time.Now().Add(WriteWait),
			); err != nil {
				c.handleErr(runtime.OpSetWriteDeadline, nil, err)
				return
			}

			if err := c.ws.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				c.handleErr(runtime.OpWritePing, nil, err)
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Conn) readPump() {
	defer c.Close()

	c.ws.SetPongHandler(func(string) error {
		if err := c.ws.SetReadDeadline(
			time.Now().Add(PongWait),
		); err != nil {
			c.handleErr(runtime.OpReadPong, nil, err)
			return err
		}

		return nil
	})

	if err := c.ws.SetReadDeadline(
		time.Now().Add(PongWait),
	); err != nil {
		c.handleErr(runtime.OpSetReadDeadline, nil, err)
		return
	}

	for {
		msgType, data, err := c.ws.ReadMessage()
		if err != nil {
			c.handleErr(runtime.OpRead, nil, err)
			return
		}

		if msgType != websocket.TextMessage {
			c.handleErr(
				runtime.OpRead,
				nil,
				ErrInvalidFrameType,
			)
			continue
		}

		if c.OnMessage != nil {
			if err := c.OnMessage(data); err != nil {
				c.handleErr(runtime.OpDecode, data, err)
			}
		}
	}
}

func (c *Conn) write(data []byte) error {
	return c.ws.WriteMessage(
		websocket.TextMessage,
		data,
	)
}

func (c *Conn) handleErr(op runtime.Op, payload any, err error) {
	log.Printf(
		"[ws][%s][%s] %v %v",
		c.ID,
		op,
		payload,
		err,
	)
}
