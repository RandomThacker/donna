package idgen

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// NewUUIDv7 returns a time-sortable UUIDv7.
func NewUUIDv7() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate uuid v7: %w", err)
	}
	return id, nil
}

// PublicID builds a prefixed public identifier from a UUID (hex without dashes).
func PublicID(prefix string, id uuid.UUID) string {
	return prefix + strings.ReplaceAll(id.String(), "-", "")
}
