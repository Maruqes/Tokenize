package StripeFunctions

import (
	"fmt"
	"net/http"

	"github.com/stripe/stripe-go/v81"
)

func callCallBack(event stripe.Event) {
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

func Custumer_subscription_deleted(w http.ResponseWriter, event stripe.Event) {
	fmt.Println("custumer_subscription_deleted")
	callCallBack(event)
}

func Customer_subscription_created(w http.ResponseWriter, event stripe.Event) {
	fmt.Println("customer_subscription_created")
	callCallBack(event)
}

func Customer_created(w http.ResponseWriter, event stripe.Event) {
	fmt.Println("customer_created")
	callCallBack(event)
}

func Invoice_payment_succeeded(w http.ResponseWriter, event stripe.Event) {
	fmt.Println("invoice_payment_succeeded")
	callCallBack(event)
}

func Charge_succeeded(w http.ResponseWriter, event stripe.Event) {
	fmt.Println("charge_succeeded")
	callCallBack(event)
}

func Invoice_created(w http.ResponseWriter, event stripe.Event) {
	fmt.Println("invoice_created")
	callCallBack(event)
}
