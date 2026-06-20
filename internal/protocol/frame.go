// Package protocol
package protocol

import "encoding/json"

// Frame represents a single WebSocket protocol message containing JSON data.
type Frame struct {
	Data json.RawMessage `json:"data,omitempty"`
}
