package Tokenize

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/price"
)

func getCustomer(id string) (*stripe.Customer, error) {
	customer, err := customer.Get(id, nil)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, err
	}

	return customer, nil
}

func checkIfIDBeingUsedInStripe(id string) bool {
	params := &stripe.CustomerListParams{}
	i := customer.List(params)
	for i.Next() {
		if i.Customer().Metadata["tokenize_id"] == id {
			return true
		}
	}
	return false
}

// if custumer already exists in stripe it does not create a new one and uses the existing one
func handleCreatingCustomer(usr database.User, customer_id string) (*stripe.Customer, error) {
	// Criar ou atualizar cliente
	finalCustomer := &stripe.Customer{}
	customerParams := &stripe.CustomerParams{
		Email: stripe.String(usr.Email),
		Metadata: map[string]string{
			"tokenize_id": customer_id,
			"username":    usr.Name,
		},
	}

	customer_exists, err := customer.Get(usr.StripeID, nil)
	if err != nil {
		log.Printf("customer.Get problem assuming it does not exists")

		if checkIfEmailIsBeingUsedInStripe(usr.Email) {
			log.Printf("email already in use")
			return nil, fmt.Errorf("email already in use")
		}

		if checkIfIDBeingUsedInStripe(customer_id) {
			log.Printf("id already in use")
			return nil, fmt.Errorf("id already in use BIG PROBLEM")
		}

		finalCustomer, err = customer.New(customerParams)
		if err != nil {
			log.Printf("customer.New: %v", err)
			return nil, err
		}
		database.SetUserStripeID(usr.ID, finalCustomer.ID)

	} else {
		finalCustomer = customer_exists
		if finalCustomer.Metadata["tokenize_id"] != customer_id {
			customerParams := &stripe.CustomerParams{
				Metadata: map[string]string{
					"tokenize_id": customer_id,
					"username":    usr.Name,
				},
			}
			_, err := customer.Update(finalCustomer.ID, customerParams)
			if err != nil {
				log.Printf("Error updating customer metadata: %v", err)
				return nil, err
			}
		}
	}

	return finalCustomer, nil
}

func validateUserInfoToActivate(customer_id string) (database.User, error) {
	customerIDInt, err := strconv.Atoi(customer_id)
	if err != nil {
		return database.User{}, fmt.Errorf("invalid user id")
	}
	exists, err := database.CheckIfUserIDExists(customerIDInt)
	if customer_id == "" || err != nil || !exists {
		return database.User{}, fmt.Errorf("invalid user id")
	}

	usr, err := database.GetUser(customerIDInt)
	if err != nil {
		return database.User{}, fmt.Errorf("error getting user")
	}
	if usr.IsActive {
		return database.User{}, fmt.Errorf("user is already active")
	}

	return usr, nil
}

func getFixedBillingCycleAnchor(month int, day int) int64 {

	now := time.Now()
	year := now.Year()
	if now.Month() > time.Month(month) || (now.Month() == time.Month(month) && now.Day() > day) {
		year++ // Caso já tenhamos passado a data fixa deste ano, avança para o próximo ano
	}
	fixedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return fixedDate.Unix()
}

func getFixedBillingFromENV() int64 {
	billing := os.Getenv("STARTING_DATE")
	day := strings.Split(billing, "/")[0]
	month := strings.Split(billing, "/")[1]

	dayInt, err := strconv.Atoi(day)
	if err != nil {
		log.Fatal("Invalid billing date")
	}
	monthInt, err := strconv.Atoi(month)
	if err != nil {
		log.Fatal("Invalid billing date")
	}

	if dayInt < 1 || dayInt > 31 || monthInt < 1 || monthInt > 12 {
		return 0
	}

	return getFixedBillingCycleAnchor(monthInt, dayInt)
}

/*
card, acss_debit, affirm, afterpay_clearpay, alipay, au_becs_debit,
bacs_debit, bancontact, blik, boleto, cashapp, customer_balance, eps,
fpx, giropay, grabpay, ideal, klarna, konbini, link, multibanco, oxxo,
p24, paynow, paypal, pix, promptpay, sepa_debit, sofort, swish, us_bank_account,
wechat_pay, revolut_pay, mobilepay, zip, amazon_pay, alma, twint, kr_card,
naver_pay, kakao_pay, payco, or samsung_pay"
*/

func createCheckoutStruct(finalCustomer *stripe.Customer) *stripe.CheckoutSessionParams {
	if GLOBAL_TYPE_OF_SUBSCRIPTION == TypeOfSubscriptionValues.OnlyStartOnDayXNoSubscription {
		panic("This function should not be called with this type of subscription")
	}

	if GLOBAL_TYPE_OF_SUBSCRIPTION == TypeOfSubscriptionValues.OnlyStartOnDayX {
		return &stripe.CheckoutSessionParams{
			Customer: stripe.String(finalCustomer.ID),
			Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")),
					Quantity: stripe.Int64(1),
				},
			},
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				BillingCycleAnchor: stripe.Int64(getFixedBillingFromENV()), // Função personalizada para calcular o timestamp
			},

			PaymentMethodTypes: stripe.StringSlice([]string{
				"card", // Standard credit/debit card payment method
			}),
			SuccessURL: stripe.String(domain + success_path),
			CancelURL:  stripe.String(domain + cancel_path),
		}
	}

	return &stripe.CheckoutSessionParams{
		Customer: stripe.String(finalCustomer.ID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")),
				Quantity: stripe.Int64(1),
			},
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card", // Standard credit/debit card payment method
		}),

		SuccessURL: stripe.String(domain + success_path),
		CancelURL:  stripe.String(domain + cancel_path),
	}
}

func createCheckoutSession(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

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

	// Configurar sessão de checkout com o cliente criado
	checkoutParams := createCheckoutStruct(finalCustomer)

	if getFixedBillingFromENV() == 0 {
		checkoutParams.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			BillingCycleAnchor: nil,
		}
	}

	s, err := session.New(checkoutParams)
	if err != nil {
		log.Printf("session.New: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	logMessage("Checkout session created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

type PrePayment struct {
	custumerID string
	date       time.Time
	type_of    string
}

var pagamentos_map = map[string]PrePayment{}

func paymentToCreateSubscriptionXDay(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

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
		Customer:           stripe.String(finalCustomer.ID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),                     // Métodos de pagamento permitidos
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)), // "Payment" para um único pagamento
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(getStringForSubscription()),
					},
					UnitAmount: &p.UnitAmount, // Valor em cêntimos
				},
				Quantity: stripe.Int64(1), // Quantidade (1 item)
			},
		},

		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			SetupFutureUsage: stripe.String("off_session"),
			Metadata: map[string]string{
				"purpose":  "Initial Subscription Payment",
				"user_id":  strconv.Itoa(customerIDInt),
				"order_id": uuid,
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

	prePayment := PrePayment{
		custumerID: customer_id,
		date:       time.Now(),
		type_of:    "Initial Subscription Payment"}
	pagamentos_map[uuid] = prePayment
}

func mourosSubscription(w http.ResponseWriter, r *http.Request) {

	//if we are inside the date range
	if checkMourosDate() {
		createCheckoutSession(w, r)
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

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
		Customer:           stripe.String(finalCustomer.ID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),                     // Métodos de pagamento permitidos
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)), // "Payment" para um único pagamento
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(getStringForSubscription()),
					},
					UnitAmount: &p.UnitAmount, // Valor em cêntimos
				},
				Quantity: stripe.Int64(1), // Quantidade (1 item)
			},
		},

		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			SetupFutureUsage: stripe.String("off_session"),
			Metadata: map[string]string{
				"purpose":  "Initial Subscription Payment Start Today",
				"user_id":  strconv.Itoa(customerIDInt),
				"order_id": uuid,
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

	prePayment := PrePayment{
		custumerID: customer_id,
		date:       time.Now(),
		type_of:    "Initial Subscription Payment Start Today"}
	pagamentos_map[uuid] = prePayment
}
