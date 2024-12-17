package Tokenize

import (
	"fmt"
	"log"
	"os"
	"strconv"

	checkouts "github.com/Maruqes/Tokenize/Checkouts"
	"github.com/Maruqes/Tokenize/Logs"
	mourosSub "github.com/Maruqes/Tokenize/MourosSub"
	types "github.com/Maruqes/Tokenize/Types"
	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
)

func handleSubscriptionDeleted(subscription stripe.Subscription) {
	fmt.Printf("Subscription deleted for %s.", subscription.ID)
	Logs.LogMessage(fmt.Sprintf("Subscription deleted for %s.", subscription.ID))
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
	customer, err := checkouts.GetCustomer(invoice.Customer.ID)
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

	//this is mouros specific
	if invoice.AmountPaid != 0 && types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.MourosSubscription {
		err := mourosSub.SetCoupon(invoice.Subscription.ID)
		if err != nil {
			log.Printf("Error setting coupon: %v", err)
			return fmt.Errorf("error setting coupon call services")
		}
	}

	fmt.Printf("\n\nSubscription PRICE ID: %s\n", subscriptionID)
	fmt.Printf("Payment succeeded for customer %s\n", invoice.Customer.ID)
	fmt.Printf("Amount: %d\n", invoice.AmountPaid)
	fmt.Printf("Product: %s\n\n\n", invoice.Lines.Data[0].Price.Product.ID)
	Logs.LogMessage(fmt.Sprintf("\n\nSubscription ID: %s", subscriptionID))
	Logs.LogMessage(fmt.Sprintf("\n\nPayment succeeded for customer %s", invoice.Customer.ID))
	Logs.LogMessage(fmt.Sprintf("Amount: %d", invoice.AmountPaid))
	Logs.LogMessage(fmt.Sprintf("Product: %s\n\n\n", invoice.Lines.Data[0].Price.Product.ID))

	return nil
}
