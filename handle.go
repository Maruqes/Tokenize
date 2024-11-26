package Tokenize

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
)

func handleSubscriptionDeleted(subscription stripe.Subscription) {
	fmt.Printf("Subscription deleted for %s.", subscription.ID)
	logMessage(fmt.Sprintf("Subscription deleted for %s.", subscription.ID))
	database.DeactivateUserByStripeID(subscription.Customer.ID)
}

func handlePaymentSuccess(invoice stripe.Invoice) error {

	subscriptionID := ""
	for _, line := range invoice.Lines.Data {
		if line.Price.ID == os.Getenv("SUBSCRIPTION_PRICE_ID") {
			subscriptionID = line.Price.ID
			break
		}
	}
	fmt.Printf("Subscription ID: %s\n", subscriptionID)

	if subscriptionID == "" {
		log.Printf("No subscription found for invoice %s", invoice.ID)
		_, err := fmt.Fprintf(os.Stderr, "No subscription found for invoice %s", invoice.ID)
		return err
	}

	fmt.Printf("Payment succeeded for customer %s\n", invoice.Customer.ID)
	fmt.Printf("Amount: %d\n", invoice.AmountPaid)
	fmt.Printf("Product: %s\n", invoice.Lines.Data[0].Price.Product.ID)

	// Recuperar metadata do cliente associado
	customer, err := getCustomer(invoice.Customer.ID)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return err
	}

	//check if the customer has the metadata
	_, exists := customer.Metadata["tokenize_id"]
	if !exists {
		log.Printf("tokenize_id metadata is missing for customer %s", customer.ID)
		return fmt.Errorf("missing tokenize_id in customer metadata")
	}
	fmt.Printf("Metadata: %v\n", customer.Metadata["tokenize_id"])

	tokenizeID, err := strconv.Atoi(customer.Metadata["tokenize_id"])
	if err != nil {
		log.Printf("Error converting tokenize_id to int: %v", err)
		return err
	}

	if err := database.SetUserStripeID(tokenizeID, invoice.Customer.ID); err != nil {
		log.Printf("Error setting user Stripe ID: %v", err)
		return err
	}

	if err := database.ActivateUser(tokenizeID); err != nil {
		log.Printf("Error activating user: %v", err)
		return err
	}
	return nil
}

func DoesUserHaveActiveSubscription(tokenizeID int) (bool, error) {
	usr, err := database.GetUser(tokenizeID)
	if err != nil {
		return false, err
	}

	if usr.IsActive {
		return true, nil
	}

	if val, err := doesHaveOfflinePayments(tokenizeID); err == nil && val {
		return true, nil
	}

	return false, nil
}

func GetUserTokenizeCookies(r *http.Request) (int, string, error) {
	cookie, err := r.Cookie("id")
	if err != nil {
		return 0, "", err
	}
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return 0, "", err
	}
	cookie, err = r.Cookie("token")
	if err != nil {
		return 0, "", err
	}
	token := cookie.Value
	return id, token, nil
}

func calculateTrialEnd(startDate int64) int64 {
	// Convert the start date from UNIX timestamp to time.Time
	startTime := time.Unix(startDate, 0)

	number_of_months := os.Getenv("NUMBER_OF_SUBSCRIPTIONS_MONTHS")
	number_of_months_int, err := strconv.Atoi(number_of_months)
	if err != nil {
		log.Printf("Error converting number of months to int: %v", err)
		return 0
	}
	// Add 1 year to the start date (adjust for the subscription duration if needed)
	trialEndTime := startTime.AddDate(0, number_of_months_int, 0) // Add 1 year

	// Convert back to UNIX timestamp
	return trialEndTime.Unix()
}

func handleInitialSubscriptionPayment(charge stripe.Charge) error {
	purpose := charge.Metadata["purpose"]
	userID := charge.Metadata["user_id"]
	orderID := charge.Metadata["order_id"]

	if purpose == "" || userID == "" || orderID == "" {
		log.Printf("Missing metadata in charge %s", charge.ID)
		return fmt.Errorf("missing metadata in charge %s", charge.ID)
	}

	userConfirm := pagamentos_map[orderID]

	if userConfirm != userID {
		log.Printf("User not found in map")
		return fmt.Errorf("user not found in map")
	}

	log.Println("Payment succeeded for user", userID)
	logMessage(fmt.Sprintf("Payment succeeded for user %s", userID))

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Error converting userID to int: %v", err)
		return err
	}
	db_user, err := database.GetUser(userIDInt)
	if err != nil {
		log.Printf("Error getting user: %v", err)
	}

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(db_user.StripeID),
		StartDate: stripe.Int64(getFixedBillingFromENV()), // Future start date in UNIX timestamp
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(1),
					},
				},
				// Set TrialEnd to the end of the first subscription period
				TrialEnd: stripe.Int64(calculateTrialEnd(getFixedBillingFromENV())),
			},
		},
	}
	schedule, err := subscriptionschedule.New(scheduleParams)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		logMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return err
	}
	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	logMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))

	return nil
}
