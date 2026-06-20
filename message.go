package cira

import (
	"encoding/json"
)

// Message represents a WebSocket protocol message with a route, type, and payload.
type Message struct {
	ID      string          `json:"id"`
	Route   string          `json:"route"`
	Type    Type            `json:"type"`
	ReplyTo string          `json:"reply_to,omitempty"`
	Data    json.RawMessage `json:"data"`
}

// Type defines the message type (request, response, or push).
type Type string

const (
	// TypeReq represents a request message that expects a response.
	TypeReq Type = "request"
	// TypeResp represents a response message that replies to a request.
	TypeResp Type = "response"
	// TypePush represents a one-way push message with no expected response.
	TypePush Type = "push"
)

// String returns the string representation of the Type.
func (t Type) String() string {
	return string(t)
}

func newMessage(route string, id string, tp Type, data json.RawMessage) *Message {
	return &Message{
		Route: route,
		ID:    id,
		Type:  tp,
		Data:  data,
	}
}
