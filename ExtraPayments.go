package Tokenize

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/price"
)

type ExtraPrePayments struct {
	custumerID         string
	date               time.Time
	type_of            string
	extra_type         string
	number_of_payments int
}

var extra_pagamentos_map = map[string]ExtraPrePayments{}

// does not work becouse stirpe does not support mbway for now
func mbwaySubscription(w http.ResponseWriter, r *http.Request) {

	if checkMourosDate() {
		createCheckoutSession(w, r)
		return
	}

	// if r.Method != "POST" {
	// 	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	// 	return
	// }

	login := CheckToken(r)
	if !login {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	//get id
	customer_id_cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Error getting id", http.StatusInternalServerError)
		return
	}
	customer_id := customer_id_cookie.Value

	//validate user
	usr, err := validateUserInfoToActivate(customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//if has offline payments
	customerIDInt, err := strconv.Atoi(customer_id)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	if val, err := doesHaveOfflinePayments(customerIDInt); err == nil && val {
		http.Error(w, "User has offline payments", http.StatusBadRequest)
		return
	}

	//creates or gets the customer
	finalCustomer, err := handleCreatingCustomer(usr, customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p, err := price.Get(os.Getenv("SUBSCRIPTION_PRICE_ID"), nil)
	if err != nil {
		http.Error(w, "Failed to get price", http.StatusInternalServerError)
		return
	}

	uuid := generateUUID()

	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(finalCustomer.ID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
			"mb_way",
		}), Mode: stripe.String(string(stripe.CheckoutSessionModePayment)), // "Payment" para um único pagamento
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Pagamento inicial para subscrição. A sua subscricão vai começar no hoje "),
					},
					UnitAmount: &p.UnitAmount, // Valor em cêntimos
				},
				Quantity: stripe.Int64(1), // Quantidade (1 item)
			},
		},

		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"purpose":    "ExtraPayExtra",
				"user_id":    strconv.Itoa(customerIDInt),
				"order_id":   uuid,
				"extra_type": "mbway",
			},
		},

		SuccessURL: stripe.String(domain + success_path),
		CancelURL:  stripe.String(domain + cancel_path),
	}

	// Cria a Checkout Session
	session, err := session.New(checkoutParams)
	if err != nil {
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	// URL para redirecionar o utilizador
	fmt.Printf("Redireciona o utilizador para: %s\n", session.URL)
	http.Redirect(w, r, session.URL, http.StatusSeeOther)
	log.Println("Payment to create subscription created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)
	logMessage("Payment to create subscription created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)

	extra_pagamentos_map[uuid] = ExtraPrePayments{
		custumerID:         customer_id,
		date:               time.Now(),
		type_of:            "ExtraPayExtraay",
		extra_type:         "mbway",
		number_of_payments: 1,
	}
}

func multibancoSubscription(w http.ResponseWriter, r *http.Request) {

	if checkMourosDate() {
		createCheckoutSession(w, r)
		return
	}

	// if r.Method != "POST" {
	// 	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	// 	return
	// }

	login := CheckToken(r)
	if !login {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	//get id
	customer_id_cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Error getting id", http.StatusInternalServerError)
		return
	}
	customer_id := customer_id_cookie.Value

	//validate user
	usr, err := validateUserInfoToActivate(customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//if has offline payments
	customerIDInt, err := strconv.Atoi(customer_id)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	if val, err := doesHaveOfflinePayments(customerIDInt); err == nil && val {
		http.Error(w, "User has offline payments", http.StatusBadRequest)
		return
	}

	//creates or gets the customer
	finalCustomer, err := handleCreatingCustomer(usr, customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p, err := price.Get(os.Getenv("SUBSCRIPTION_PRICE_ID"), nil)
	if err != nil {
		http.Error(w, "Failed to get price", http.StatusInternalServerError)
		return
	}

	uuid := generateUUID()

	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(finalCustomer.ID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
			"multibanco",
		}), Mode: stripe.String(string(stripe.CheckoutSessionModePayment)), // "Payment" para um único pagamento
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Pagamento inicial para subscrição. A sua subscricão vai começar no hoje "),
					},
					UnitAmount: &p.UnitAmount, // Valor em cêntimos
				},
				Quantity: stripe.Int64(1), // Quantidade (1 item)
			},
		},

		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"purpose":    "ExtraPayExtra",
				"user_id":    strconv.Itoa(customerIDInt),
				"order_id":   uuid,
				"extra_type": "multibanco",
			},
		},

		SuccessURL: stripe.String(domain + success_path),
		CancelURL:  stripe.String(domain + cancel_path),
	}

	// Cria a Checkout Session
	session, err := session.New(checkoutParams)
	if err != nil {
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	// URL para redirecionar o utilizador
	fmt.Printf("Redireciona o utilizador para: %s\n", session.URL)
	http.Redirect(w, r, session.URL, http.StatusSeeOther)
	log.Println("Payment to create subscription created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)
	logMessage("Payment to create subscription created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)

	extra_pagamentos_map[uuid] = ExtraPrePayments{
		custumerID:         customer_id,
		date:               time.Now(),
		type_of:            "ExtraPayExtra",
		extra_type:         "multibanco",
		number_of_payments: 1,
	}
}
