package functions

import (
	"log"
	"net/mail"
	"os"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	uuid := uuid.New() // Generate a new UUIDv4
	return uuid.String()
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func CheckAllEnv() {
	requiredEnvVars := []string{
		"SECRET_KEY",
		"ENDPOINT_SECRET",
		"SUBSCRIPTION_PRICE_ID",
		"DOMAIN",
		"LOGS_FILE",
		"SECRET_ADMIN",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Missing env variable: %s", envVar)
		}
	}
}
