package Tokenize

import (
	"net/mail"

	"github.com/google/uuid"
)

func generateUUID() string {
	uuid := uuid.New() // Generate a new UUIDv4
	return uuid.String()
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
