package Tokenize

import (
	"github.com/google/uuid"
)

func generateUUID() string {
	uuid := uuid.New() // Generate a new UUIDv4
	return uuid.String()
}
