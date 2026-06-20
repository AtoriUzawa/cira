package cira

import (
	"errors"
	"time"

	"github.com/AtoriUzawa/cira/internal/idgen"
	"github.com/AtoriUzawa/cira/internal/protocol"
)

// Context holds the state for handling a single WebSocket message, including the client, message, and codec.
type Context struct {
	ClientID string
	Message  *Message
	Timeout  time.Duration
	Conn     *Conn

	client *client

	idGenerator idgen.IDGenerator
	codec       protocol.Codec
}

const defaultTimeout = 10 * time.Hour

// ErrCallTimeout is returned when a Call request exceeds the configured timeout.
var ErrCallTimeout = errors.New("ws/context: call timeout")

func newContext(client *client, idGenerator idgen.IDGenerator, codec protocol.Codec) *Context {
	ctx := &Context{
		Timeout:     defaultTimeout,
		client:      client,
		idGenerator: idGenerator,
		codec:       codec,
		ClientID:    client.ID(),
	}
	return ctx
}

// Push sends a push-type event with the given data to the client.
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

	c.client.Send(b)

	return nil
}

// Req sends a request-type event with the given data to the client.
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

	c.client.Send(b)

	return nil
}

// Resp sends a response to the client, correlating with the current request message.
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

	c.client.Send(b)
}

// Call sends a request to the client and waits for a response, returning an error if the timeout is exceeded.
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

	ch := c.client.Runtime.Register(msg.ID)

	c.client.Send(b)

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
		c.client.Runtime.Unregister(msg.ID)
		return ErrCallTimeout
	}
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
