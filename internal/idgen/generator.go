// Package idgen
package idgen

import (
	"github.com/google/uuid"
)

// IDGenerator generates unique string identifiers.
type IDGenerator interface {
	Next() string
}

// UUIDGenerator is an IDGenerator that generates UUID v4 strings.
type UUIDGenerator struct{}

// Next returns a new UUID v4 string.
func (g *UUIDGenerator) Next() string {
	return uuid.NewString()
}
