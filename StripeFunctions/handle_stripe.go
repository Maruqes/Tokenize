package StripeFunctions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
)

func CallCallBack(event stripe.Event) {
	//get callback metadata from event
	metadata, ok := event.Data.Object["metadata"].(map[string]interface{})
	if ok {
		callbackVal, ok := metadata["callback"]
		if ok {
			if callbackStr, ok := callbackVal.(string); ok {
				fmt.Println("callbackID:", callbackStr)
				callback := getCallback(callbackStr)
				if callback != nil {
					callback(event)
				}
			}
		}
	}
}

func Customer_created(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	fmt.Println("customer_created")
}

func Custumer_subscription_deleted(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	fmt.Println("custumer_subscription_deleted")
	// active -> 0
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := database.GetUserByStripeID(subscription.Customer.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !user.IsActive {
		fmt.Println("user is already inactive")
		return
	}

	err = database.DeactivateUser(user.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func Customer_subscription_created(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	fmt.Println("customer_subscription_created")
}

func Invoice_payment_succeeded(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	fmt.Println("invoice_payment_succeeded")
	// active -> 1
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := database.GetUserByStripeID(invoice.Customer.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if user.IsActive {
		fmt.Println("user is already active")
		return
	}

	err = database.ActivateUser(user.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func Charge_succeeded(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	fmt.Println("charge_succeeded")
}

func Invoice_created(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	fmt.Println("invoice_created")
}
