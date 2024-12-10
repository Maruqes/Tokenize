package functions

import (
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	types "github.com/Maruqes/Tokenize/Types"
	"github.com/Maruqes/Tokenize/database"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
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

	IsValidDate(os.Getenv("STARTING_DATE"))
	IsValidDate(os.Getenv("MOUROS_STARTING_DATE"))
	IsValidDate(os.Getenv("MOUROS_ENDING_DATE"))
}

func IsValidDate(dates string) {
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

func HasStartingDayPassed() bool {
	startingDate := os.Getenv("STARTING_DATE")
	startingDay := strings.Split(startingDate, "/")[0]
	startingMonth := strings.Split(startingDate, "/")[1]

	startingDayInt, err := strconv.Atoi(startingDay)
	if err != nil {
		log.Fatal("Invalid day format")
	}
	startingMonthInt, err := strconv.Atoi(startingMonth)
	if err != nil {
		log.Fatal("Invalid month format")
	}

	now := time.Now()
	startingDate_date := time.Date(now.Year(), time.Month(startingMonthInt), startingDayInt, 0, 0, 0, 0, time.UTC)

	return now.After(startingDate_date)
}

func CheckMourosDate() bool {
	mourosStartDate := os.Getenv("MOUROS_STARTING_DATE")
	mourosEndDate := os.Getenv("MOUROS_ENDING_DATE")

	if mourosStartDate == "" || mourosEndDate == "" {
		return false
	}

	// Parse the dates in day/month format
	startingDateParts := strings.Split(mourosStartDate, "/")
	endingDateParts := strings.Split(mourosEndDate, "/")

	if len(startingDateParts) != 2 || len(endingDateParts) != 2 {
		return false
	}

	startingDay, err := strconv.Atoi(startingDateParts[0])
	if err != nil {
		return false
	}
	startingMonth, err := strconv.Atoi(startingDateParts[1])
	if err != nil {
		return false
	}

	endingDay, err := strconv.Atoi(endingDateParts[0])
	if err != nil {
		return false
	}
	endingMonth, err := strconv.Atoi(endingDateParts[1])
	if err != nil {
		return false
	}

	now := time.Now()
	startingDate := time.Date(now.Year(), time.Month(startingMonth), startingDay, 0, 0, 0, 0, time.UTC)
	endingDate := time.Date(now.Year(), time.Month(endingMonth), endingDay, 23, 59, 59, 0, time.UTC)

	if now.After(startingDate) && now.Before(endingDate) {
		return true
	}

	return false
}

func DoesUserHaveActiveSubscription(tokenizeID int) (bool, error) {
	usr, err := database.GetUser(tokenizeID)
	if err != nil {
		return false, err
	}

	if usr.IsActive {
		return true, nil
	}

	return false, nil
}

func GetStringForSubscription() string {
	if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.Normal {
		return "Your subscription will start today"
	} else if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.OnlyStartOnDayX {
		return "Your subscription will start today"
	} else if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.OnlyStartOnDayXNoSubscription {
		month_day := os.Getenv("STARTING_DATE")
		monthStr := strings.Split(month_day, "/")[1]
		dayStr := strings.Split(month_day, "/")[0]

		month, err := strconv.Atoi(monthStr)
		if err != nil {
			log.Fatal("Invalid month format")
		}
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			log.Fatal("Invalid day format")
		}

		starting_date := time.Date(time.Now().Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if time.Now().After(starting_date) {
			starting_date = time.Date(time.Now().Year()+1, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
		return "Your subscription will start on " + starting_date.Format("02/01/2006")
	} else if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.MourosSubscription {
		return "Your subscription will start today"
	}
	return ""
}

func DefinePaymentMethod(customerID string, paymentIntentID string) error {
	paymentIntent, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		log.Printf("Erro ao obter o PaymentIntent: %v", err)
		return err
	}

	lastPaymentMethodID := paymentIntent.PaymentMethod.ID

	customerUpdateParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(lastPaymentMethodID),
		},
	}
	_, err = customer.Update(customerID, customerUpdateParams)
	if err != nil {
		log.Printf("Erro ao definir o método de pagamento padrão: %v", err)
		return err
	}
	return nil
}