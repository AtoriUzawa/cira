package cira

import (
	"errors"
	"time"

	"github.com/AtoriUzawa/cira/internal/runtime"
)

type Stream struct {
	id   string
	ctx  *Context
	data <-chan *runtime.Delivery
}

var ErrStreamTimeout = errors.New("ws/stream: stream timeout")

func (s *Stream) Send(data any) error {
	b, err := s.ctx.codec.Encode(data)
	if err != nil {
		return err
	}

	msg := &Message{
		Type:    TypeStream,
		ReplyTo: s.id,
		Data:    b,
	}

	b, err = s.ctx.codec.Encode(msg)
	if err != nil {
		return err
	}

	s.ctx.peer.Send(b)

	return nil
}

func (s *Stream) Recv(resp any) error {
	result := <-s.data
	b := result.Data

	if b == nil {
		return result.Err
	}

	if err := s.ctx.codec.Decode(b, resp); err != nil {
		return err
	}

	return nil
}

func (s *Stream) RecvTimeout(resp any) error {
	timer := time.NewTimer(s.ctx.Timeout)
	defer timer.Stop()

	select {
	case result := <-s.data:
		b := result.Data

		if b == nil {
			return result.Err
		}

		if err := s.ctx.codec.Decode(b, resp); err != nil {
			return err
		}
	case <-timer.C:
		return ErrStreamTimeout
	}

	return nil
}
