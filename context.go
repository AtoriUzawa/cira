package cira

import (
	"errors"
	"time"

	"github.com/AtoriUzawa/cira/internal/idgen"
	"github.com/AtoriUzawa/cira/internal/protocol"
)

// Context holds the state for handling a single WebSocket message, including the peer, message, and codec.
type Context struct {
	PeerID  string
	Message *Message
	Timeout time.Duration
	Conn    *Conn

	peer   *peer
	stream *Stream

	idGenerator idgen.IDGenerator
	codec       protocol.Codec
}

const defaultTimeout = 30 * time.Second

// ErrCallTimeout is returned when a Call request exceeds the configured timeout.
var ErrCallTimeout = errors.New("ws/context: call timeout")

func newContext(peer *peer, idGenerator idgen.IDGenerator, codec protocol.Codec) *Context {
	ctx := &Context{
		Timeout:     defaultTimeout,
		peer:        peer,
		idGenerator: idGenerator,
		codec:       codec,
		PeerID:      peer.ID(),
	}
	return ctx
}

// Push sends a push-type event with the given data to the peer.
func (c *Context) Push(event string, data any) error {
	b, err := c.codec.Encode(data)
	if err != nil {
		return err
	}

	msg := newMessage(event, c.idGenerator.Next(), TypePush, b)

	b, err = c.codec.Encode(msg)
	if err != nil {
		return err
	}

	c.peer.Send(b)

	return nil
}

// Req sends a request-type event with the given data to the peer.
func (c *Context) Req(event string, data any) error {
	b, err := c.codec.Encode(data)
	if err != nil {
		return err
	}

	msg := newMessage(event, c.idGenerator.Next(), TypeReq, b)

	b, err = c.codec.Encode(msg)
	if err != nil {
		return err
	}

	c.peer.Send(b)

	return nil
}

// Resp sends a response to the peer, correlating with the current request message.
func (c *Context) Resp(data any) {
	b, err := c.codec.Encode(data)
	if err != nil {
		return
	}

	m := newWithReq(c.idGenerator.Next(), c.Message, b)

	b, err = c.codec.Encode(m)
	if err != nil {
		return
	}

	c.peer.Send(b)
}

// Call sends a request to the peer and waits for a response, returning an error if the timeout is exceeded.
func (c *Context) Call(route string, req any, resp any) error {
	b, err := c.codec.Encode(req)
	if err != nil {
		return err
	}
	msg := newMessage(route, c.idGenerator.Next(), TypeReq, b)
	b, err = c.codec.Encode(msg)
	if err != nil {
		return err
	}

	ch := c.peer.Runtime.Register(msg.ID)

	c.peer.Send(b)

	timer := time.NewTimer(c.Timeout)
	defer timer.Stop()
	select {
	case result := <-ch:
		b := result.Data
		if b == nil {
			return result.Err
		}

		if err := c.codec.Decode(b, resp); err != nil {
			return err
		}

		return nil
	case <-timer.C:
		c.peer.Runtime.Unregister(msg.ID)
		return ErrCallTimeout
	}
}

func (c *Context) OpenStream(id string) *Stream {
	s := &Stream{
		id,
		c,
		c.peer.Runtime.Register(id),
	}

	c.peer.OnClose(func() {
		c.CloseStream()
	})

	c.stream = s
	return s
}

func (c *Context) CloseStream() {
	if c.stream == nil {
		return
	}

	msg := Message{
		Type:    TypeStreamClose,
		ReplyTo: c.stream.id,
	}

	b, _ := c.codec.Encode(msg)
	c.peer.Send(b)

	c.peer.Runtime.Unregister(c.stream.id)

	c.stream.id = ""
	c.stream.data = nil
	c.stream = nil
}

func newWithReq(id string, req *Message, data []byte) *Message {
	return &Message{
		Route:   req.Route,
		ID:      id,
		Type:    TypeResp,
		ReplyTo: req.ID,
		Data:    data,
	}
}
