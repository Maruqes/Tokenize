package Tokenize

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
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

	if subscriptionID == "" {
		log.Printf("No subscription found for invoice %s", invoice.ID)
		_, err := fmt.Fprintf(os.Stderr, "No subscription found for invoice %s", invoice.ID)
		return err
	}

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

	fmt.Printf("\n\nSubscription ID: %s\n", subscriptionID)
	fmt.Printf("Payment succeeded for customer %s\n", invoice.Customer.ID)
	fmt.Printf("Amount: %d\n", invoice.AmountPaid)
	fmt.Printf("Product: %s\n\n\n", invoice.Lines.Data[0].Price.Product.ID)
	logMessage(fmt.Sprintf("\n\nSubscription ID: %s", subscriptionID))
	logMessage(fmt.Sprintf("\n\nPayment succeeded for customer %s", invoice.Customer.ID))
	logMessage(fmt.Sprintf("Amount: %d", invoice.AmountPaid))
	logMessage(fmt.Sprintf("Product: %s\n\n\n", invoice.Lines.Data[0].Price.Product.ID))

	return nil
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

// Initial Subscription Payment
func handleInitialSubscriptionPayment(charge stripe.Charge) error {
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
		PanicLog("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED\n\n")
		fmt.Println("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED")
		return fmt.Errorf("if you seeing this conect support or stop messing with the requests")
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

	err = definePaymentMethod(db_user.StripeID, charge.PaymentIntent.ID)
	if err != nil {
		log.Printf("Erro ao definir o método de pagamento padrão: %v", err)
		return err
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

	delete(pagamentos_map, orderID)
	return nil
}

func firstSubscriptionMoure(userid string) (*stripe.SubscriptionSchedule, error) {
	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(userid),
		StartDate: stripe.Int64(time.Now().Unix()),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(1),
					},
				},
				TrialEnd: stripe.Int64(getFixedBillingFromENV()),
				EndDate:  stripe.Int64(getFixedBillingFromENV()),
			},
		},
		EndBehavior: stripe.String("cancel"),
	}
	schedule, err := subscriptionschedule.New(scheduleParams)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		logMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return nil, err
	}
	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	logMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))
	return schedule, nil
}

func secondSubscriptionMoure(userid string) (*stripe.SubscriptionSchedule, error) {

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(userid),
		StartDate: stripe.Int64(getFixedBillingFromENV()), // Future start date in UNIX timestamp
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(1),
					},
				},
			},
		},
	}
	schedule, err := subscriptionschedule.New(scheduleParams)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		logMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return nil, err
	}
	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	logMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))
	return schedule, nil
}

func definePaymentMethod(customerID string, paymentIntentID string) error {
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

// Initial Subscription Payment Start Today
func handleInitialSubscriptionPaymentStartToday(charge stripe.Charge) error {
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

	if userConfirm.type_of != "Initial Subscription Payment Start Today" {
		PanicLog("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED\n\n")
		fmt.Println("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED")
		return fmt.Errorf("if you seeing this conect support or stop messing with the requests")
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
		return err
	}

	err = definePaymentMethod(db_user.StripeID, charge.PaymentIntent.ID)
	if err != nil {
		log.Printf("Erro ao definir o método de pagamento padrão: %v", err)
		return err
	}

	schedule, err := firstSubscriptionMoure(db_user.StripeID)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		logMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return err
	}
	fmt.Printf("First Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	logMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))

	schedule, err = secondSubscriptionMoure(db_user.StripeID)
	if err != nil {
		log.Printf("Second Error creating subscription schedule: %v", err)
		logMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return err
	}

	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	logMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))

	//delete the order from the map
	delete(pagamentos_map, orderID)

	return nil
}
