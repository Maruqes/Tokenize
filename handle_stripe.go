package Tokenize

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v81"
)

func custumer_subscription_deleted(w http.ResponseWriter, event stripe.Event) {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Subscription deleted for %s.", subscription.ID)
	handleSubscriptionDeleted(subscription)
}

func customer_subscription_created(w http.ResponseWriter, event stripe.Event) {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Subscription created for %s.", subscription.ID)
	logMessage("Subscription in stripe created for " + subscription.Customer.ID)
}

func customer_created(w http.ResponseWriter, event stripe.Event) {
	var our_customer stripe.Customer
	err := json.Unmarshal(event.Data.Raw, &our_customer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Customer created for %s with email %s  ID:%s.", our_customer.Name, our_customer.Email, our_customer.ID)
	logMessage("Customer in stripe created for " + our_customer.ID)
}

func invoice_payment_succeeded(w http.ResponseWriter, event stripe.Event) {
	// caso pagamento de subscricao normal, pagou tem direito
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Invoice payment succeeded for %s.", invoice.ID)
	logMessage("Invoice payment succeeded for " + invoice.Customer.ID)
	err = handlePaymentSuccess(invoice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling payment success: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func charge_succeeded(w http.ResponseWriter, event stripe.Event) {
	var charge stripe.Charge
	err := json.Unmarshal(event.Data.Raw, &charge)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Charge succeeded for %s.", charge.ID)
	logMessage("Charge succeeded for " + charge.Customer.ID)
	purpose := charge.Metadata["purpose"]
	userID := charge.Metadata["user_id"]
	orderID := charge.Metadata["order_id"]

	if purpose == "Initial Subscription Payment" {
		//criar subscricao
		if err := handleInitialSubscriptionPayment(charge); err != nil {
			fmt.Fprintf(os.Stderr, "Error handling initial subscription payment: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	log.Printf("Charge succeeded for %s (Purpose: %s, UserID: %s, OrderID: %s).", charge.ID, purpose, userID, orderID)

}

func invoice_created(w http.ResponseWriter, event stripe.Event) {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Invoice payment succeeded for %s.", invoice.ID)
	logMessage("Invoice payment succeeded for " + invoice.Customer.ID)
	err = handlePaymentSuccess(invoice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling payment success: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}