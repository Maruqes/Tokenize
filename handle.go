package Tokenize

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
)

func handle_cancel_update(subscription stripe.Subscription) {
	fmt.Println(subscription.CancelAt)   //quando a atual assinatura será cancelada
	fmt.Println(subscription.CanceledAt) //quando foi cancelada
}

func handleSubscriptionDeleted(subscription stripe.Subscription) {
	fmt.Printf("Subscription deleted for %s.", subscription.ID)
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
