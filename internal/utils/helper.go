package utils

import (
	"time"

	"github.com/google/uuid"
)

// Ms returns duration in milliseconds as string
func Ms(d time.Duration) string {
	return d.String()
}

// ParseUUID parses a string into UUID
func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
