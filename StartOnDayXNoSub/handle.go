package startOnDayXNoSub

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	checkouts "github.com/Maruqes/Tokenize/Checkouts"
	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/Logs"
	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
)

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

// Initial Subscription Payment
func HandleInitialSubscriptionPayment(charge stripe.Charge) error {
	purpose := charge.Metadata["purpose"]
	userID := charge.Metadata["user_id"]
	orderID := charge.Metadata["order_id"]

	if purpose == "" || userID == "" || orderID == "" {
		log.Printf("Missing metadata in charge %s", charge.ID)
		return fmt.Errorf("missing metadata in charge %s", charge.ID)
	}

	userConfirm, exists := pagamentos_map[orderID]
	if !exists {
		log.Printf("Order ID %s not found in map", orderID)
		return fmt.Errorf("order ID %s not found in map", orderID)
	}

	if userConfirm.custumerID != userID {
		log.Printf("User not found in map")
		return fmt.Errorf("user not found in map")
	}

	if userConfirm.type_of != "Initial Subscription Payment" {
		Logs.PanicLog("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED\n\n")
		fmt.Println("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED")
		return fmt.Errorf("if you seeing this conect support or stop messing with the requests")
	}

	log.Println("Payment succeeded for user", userID)
	Logs.LogMessage(fmt.Sprintf("Payment succeeded for user %s", userID))

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Error converting userID to int: %v", err)
		return err
	}
	db_user, err := database.GetUser(userIDInt)
	if err != nil {
		log.Printf("Error getting user: %v", err)
	}

	err = functions.DefinePaymentMethod(db_user.StripeID, charge.PaymentIntent.ID)
	if err != nil {
		log.Printf("Erro ao definir o método de pagamento padrão: %v", err)
		return err
	}

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(db_user.StripeID),
		StartDate: stripe.Int64(checkouts.GetFixedBillingFromENV()), // Future start date in UNIX timestamp
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(1),
					},
				},
				// Set TrialEnd to the end of the first subscription period
				TrialEnd: stripe.Int64(calculateTrialEnd(checkouts.GetFixedBillingFromENV())),
			},
		},
	}
	schedule, err := subscriptionschedule.New(scheduleParams)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return err
	}
	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	Logs.LogMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))

	delete(pagamentos_map, orderID)
	return nil
}
