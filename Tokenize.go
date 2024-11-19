package Tokenize

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	portalsession "github.com/stripe/stripe-go/v81/billingportal/session"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/webhook"
)

//tipos de admin -> 0:superadmin, 1:admin, -1:sem acesso
//para registrar tens de ser admin 0
// tipo_admin   permissoes
// 0			//tudo porque é 0
// 1            //loja, produto

//criar conta-sistema de pagamentos

//criar conta normal  com perm -1
//form de pagamento->pagar   //dar para pagar em dinheiro
//é membro

//uma subricicao
//dar duracao a subscricao

var domain = os.Getenv("DOMAIN")
var Permissions = permissions{}

func getCustomer(id string) (*stripe.Customer, error) {
	customer, err := customer.Get(id, nil)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, err
	}

	return customer, nil
}

func checkIfEmailIsBeingUsedInStripe(email string) bool {
	params := &stripe.CustomerListParams{
		Email: stripe.String(email),
	}
	i := customer.List(params)
	for i.Next() {
		if i.Customer().Email == email {
			return true
		}
	}
	return false
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
	} else {
		finalCustomer = customer_exists
	}

	return finalCustomer, nil
}

func validateUserInfoToActivate(customer_id string) (database.User, error) {
	customerIDInt, err := strconv.Atoi(customer_id)
	if customer_id == "" || err != nil || !database.CheckIfUserIDExists(customerIDInt) {
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

func createCheckoutSession(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	login := checkToken(r)
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

	//creates or gets the customer
	finalCustomer, err := handleCreatingCustomer(usr, customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Configurar sessão de checkout com o cliente criado
	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(finalCustomer.ID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(domain + "/success.html?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(domain + "/cancel.html"),
	}

	s, err := session.New(checkoutParams)
	if err != nil {
		log.Printf("session.New: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func createPortalSession(w http.ResponseWriter, r *http.Request) {
	// For demonstration purposes, we're using the Checkout session to retrieve the customer ID.
	// Typically this is stored alongside the authenticated user in your database.

	login := checkToken(r)
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

	//get customer
	customerIDInt, err := strconv.Atoi(customer_id)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	usr, err := database.GetUser(customerIDInt)
	if err != nil {
		http.Error(w, "Error getting customer", http.StatusInternalServerError)
		return
	}

	// Authenticate your user.
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(usr.StripeID),
		ReturnURL: stripe.String(domain),
	}
	ps, _ := portalsession.New(params)
	log.Printf("ps.New: %v", ps.URL)
	http.Redirect(w, r, ps.URL, http.StatusSeeOther)
}

func handleWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	bodyReader := http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(bodyReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks
	endpointSecret := os.Getenv("ENDPOINT_SECRET")
	if endpointSecret == "" {
		fmt.Fprintf(os.Stderr, "Error reading endpoint secret: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription deleted for %s.", subscription.ID)
		handleSubscriptionDeleted(subscription)
	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription updated for %s.", subscription.ID)
		handle_cancel_update(subscription)
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription created for %s.", subscription.ID)
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionCreated(subscription)
	case "customer.subscription.trial_will_end":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription trial will end for %s.", subscription.ID)
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionTrialWillEnd(subscription)
	case "entitlements.active_entitlement_summary.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Active entitlement summary updated for %s.", subscription.ID) // Then define and call a func to handle active entitlement summary updated.
		// handleEntitlementUpdated(subscription)
	case "customer.created":
		var our_customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &our_customer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Customer created for %s with email %s  ID:%s.", our_customer.Name, our_customer.Email, our_customer.ID)
	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Invoice payment succeeded for %s.", invoice.ID)
		err = handlePaymentSuccess(invoice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error handling payment success: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Then define and call a func to handle the successful payment of an invoice.
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}
	w.WriteHeader(http.StatusOK)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	login := checkToken(r)
	if login {
		http.Error(w, "Already logged in, cant create an account", http.StatusUnauthorized)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received credentials: %s / %s", credentials.Username, credentials.Password)

	if credentials.Username == "" || credentials.Password == "" || credentials.Email == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	//nao pode criar uma conta cujo email ja esta na stripe
	if checkIfEmailIsBeingUsedInStripe(credentials.Email) {
		http.Error(w, "Email already in use", http.StatusBadRequest)
		return
	}

	//enviar email de confirmacao

	id, err := database.AddUser("", credentials.Email, credentials.Username, credentials.Password)
	if err != nil {
		http.Error(w, "Failed to create user check the credentials", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": %d}`, id)))
}

func loginUsr(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received credentials: %s / %s", credentials.Email, credentials.Password)

	if credentials.Email == "" || credentials.Password == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	token, usr, err := loginUser(credentials.Email, credentials.Password)
	if err != nil || token == "" {
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	// Set cookies with expiration in 5 days
	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    strconv.Itoa(usr.ID),
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(5 * 24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(5 * 24 * time.Hour),
	})

	w.WriteHeader(http.StatusOK)
}

func testLogin(w http.ResponseWriter, r *http.Request) {
	login_Q := checkToken(r)
	if !login_Q {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Error getting id", http.StatusInternalServerError)
		return
	}

	idInt, err := strconv.Atoi(id.Value)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	usr, err := database.GetUser(idInt)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Logged in as " + usr.Name))
}

func logoutUsr(w http.ResponseWriter, r *http.Request) {
	login_Q := checkToken(r)
	if !login_Q {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Error getting id", http.StatusInternalServerError)
		return
	}

	idInt, err := strconv.Atoi(id.Value)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	logoutUser(idInt)
	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// set port like "4242"
func Init(port string) {

	//check if env variables are set
	if os.Getenv("SECRET_KEY") == "" ||
		os.Getenv("ENDPOINT_SECRET") == "" ||
		os.Getenv("SUBSCRIPTION_PRICE_ID") == "" ||
		os.Getenv("DOMAIN") == "" {
		log.Fatal("Missing env variables")
	}

	port_int, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Invalid port")
	}
	if port_int <= 0 || port_int > 65535 {
		log.Fatal("Invalid port")
	}

	fmt.Println("Init")

	database.Init()
	database.CreateTable()
	database.CreatePermissionsTable()

	stripe.Key = os.Getenv("SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir("public"))) //for testing

	http.HandleFunc("/create-checkout-session", createCheckoutSession) //subscricao
	http.HandleFunc("/create-portal-session", createPortalSession)     //para checkar info da subscricao
	http.HandleFunc("/webhook", handleWebhook)

	//testing
	http.HandleFunc("/testeLOGIN", testLogin)

	//db
	http.HandleFunc("/create-user", createUser)

	//auth
	http.HandleFunc("/login-user", loginUsr)
	http.HandleFunc("/logout-user", logoutUsr)

	addr := "localhost:" + port
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
