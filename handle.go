package Tokenize

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
)

func handleSubscriptionDeleted(subscription stripe.Subscription) {
}

func handlePaymentSuccess(invoice stripe.Invoice) error {
	return nil
}
