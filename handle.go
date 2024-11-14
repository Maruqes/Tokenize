package Tokenize

import (
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
)

func handle_cancel_update(subscription stripe.Subscription) {
	fmt.Println(subscription.CancelAt)   //quando a atual assinatura será cancelada
	fmt.Println(subscription.CanceledAt) //quando foi cancelada
}

func handleSubscriptionCanceled(subscription stripe.Subscription) {
	fmt.Printf("Subscription deleted for %s.", subscription.ID)
}

func handle_payment_success(invoice stripe.Invoice) {
	fmt.Printf("Payment succeeded for %s for the amount of %d for product %s.", invoice.CustomerName, invoice.AmountPaid, invoice.Lines.Data[0].Price.Product.ID)
}
