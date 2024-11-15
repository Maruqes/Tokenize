package Tokenize

import (
	"Tokenize/database"
	"fmt"
	"log"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
)

func handle_cancel_update(subscription stripe.Subscription) {
	fmt.Println(subscription.CancelAt)   //quando a atual assinatura será cancelada
	fmt.Println(subscription.CanceledAt) //quando foi cancelada
}

func handleSubscriptionDeleted(subscription stripe.Subscription) {
	fmt.Printf("Subscription deleted for %s.", subscription.ID)
}

func handlePaymentSuccess(invoice stripe.Invoice) {
	fmt.Printf("Payment succeeded for customer %s\n", invoice.Customer.ID)
	fmt.Printf("Amount: %d\n", invoice.AmountPaid)
	fmt.Printf("Product: %s\n", invoice.Lines.Data[0].Price.Product.ID)

	// Recuperar metadata do cliente associado
	customer, err := getCustomer(invoice.Customer.ID)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return
	}

	fmt.Printf("Metadata: %v\n", customer.Metadata["tokenize_id"])

	tokenizeID, err := strconv.Atoi(customer.Metadata["tokenize_id"])
	if err != nil {
		log.Printf("Error converting tokenize_id to int: %v", err)
		return
	}
	database.SetUserStripeID(tokenizeID, invoice.Customer.ID)
}
