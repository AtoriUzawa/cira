package protocol

import (
	"encoding/json"
	"fmt"
)

// Codec defines the interface for encoding and decoding messages.
type Codec interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
}

// JSONCodec is a Codec implementation that uses JSON serialization.
type JSONCodec struct{}

// Encode serializes a value into a JSON byte slice.
func (JSONCodec) Encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json encode failed: %w", err)
	}

	return b, nil
}

// Decode deserializes a JSON byte slice into the provided value.
func (JSONCodec) Decode(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("json decode failed: %w", err)
	}

	return nil
}
