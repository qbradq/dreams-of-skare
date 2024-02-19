package util

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/google/uuid"
)

// NewUUID returns a new, random UUID in string representation.
func NewUUID() string {
	u, _ := uuid.NewUUID()
	return hex.EncodeToString(u[:16])
}

// Hash returns the hash of the string as a string.
func Hash(s string) string {
	sum := sha512.Sum512([]byte(s))
	return hex.EncodeToString(sum[:sha512.Size])
}
