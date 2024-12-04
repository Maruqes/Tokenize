package Tokenize

import (
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81/price"
)

func generateUUID() string {
	uuid := uuid.New() // Generate a new UUIDv4
	return uuid.String()
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func checkAllEnv() {
	requiredEnvVars := []string{
		"SECRET_KEY",
		"ENDPOINT_SECRET",
		"SUBSCRIPTION_PRICE_ID",
		"DOMAIN",
		"LOGS_FILE",
		"SECRET_ADMIN",
		"NUMBER_OF_SUBSCRIPTIONS_MONTHS",
		"STARTING_DATE",
		"MOUROS_STARTING_DATE",
		"MOUROS_ENDING_DATE",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Missing env variable: %s", envVar)
		}
	}

	if !isValidPriceID(os.Getenv("SUBSCRIPTION_PRICE_ID")) {
		log.Fatal("Invalid SUBSCRIPTION_PRICE_ID")
	}

	isValidDate(os.Getenv("STARTING_DATE"))
	isValidDate(os.Getenv("MOUROS_STARTING_DATE"))
	isValidDate(os.Getenv("MOUROS_ENDING_DATE"))
}

func isValidPriceID(priceID string) bool {
	_, err := price.Get(priceID, nil)

	return err == nil
}

func isValidDate(dates string) {
	day := strings.Split(dates, "/")[0]
	month := strings.Split(dates, "/")[1]

	dayInt, err := strconv.Atoi(day)
	if err != nil {
		log.Fatal("Invalid day format")
	}
	monthInt, err := strconv.Atoi(month)
	if err != nil {
		log.Fatal("Invalid month format")
	}
	if dayInt > 31 || monthInt > 12 || dayInt < 1 || monthInt < 1 {
		log.Fatal("Invalid date")
	}

}
