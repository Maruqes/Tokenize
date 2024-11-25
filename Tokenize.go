package Tokenize

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

var success_path = ""
var cancel_path = ""

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
	checkoutParams := &stripe.CheckoutSessionParams{
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

		SuccessURL: stripe.String(domain + success_path),
		CancelURL:  stripe.String(domain + cancel_path),
	}

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

func createPortalSession(w http.ResponseWriter, r *http.Request) {
	// For demonstration purposes, we're using the Checkout session to retrieve the customer ID.
	// Typically this is stored alongside the authenticated user in your database.

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
	logMessage("Portal session created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)
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
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription created for %s.", subscription.ID)
		logMessage("Subscription in stripe created for " + subscription.Customer.ID)
	case "customer.created":
		var our_customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &our_customer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Customer created for %s with email %s  ID:%s.", our_customer.Name, our_customer.Email, our_customer.ID)
		logMessage("Customer in stripe created for " + our_customer.ID)
	case "invoice.payment_succeeded":
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

	login := CheckToken(r)
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

	logMessage("Create user attempt with email " + credentials.Email + " and username " + credentials.Username)

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
		http.Error(w, "Failed to create user check the credentials with err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	logMessage("User created with id/name " + strconv.Itoa(int(id)) + "/" + credentials.Username)

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

	logMessage("Login attempt with email " + credentials.Email)

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

	logMessage("User logged in with id/name " + strconv.Itoa(usr.ID) + "/" + usr.Name)

	w.WriteHeader(http.StatusOK)
}

func testLogin(w http.ResponseWriter, r *http.Request) {
	login_Q := CheckToken(r)
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
	login_Q := CheckToken(r)
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

	logMessage("User logged out with id " + strconv.Itoa(idInt))

	w.WriteHeader(http.StatusOK)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// set port like "4242"
func Init(port string, success string, cancel string) {

	success_path = success
	cancel_path = cancel

	if success[0] != '/' || cancel[0] != '/' {
		panic("Success/Cancel path must start with /")
	}

	//check if env variables are set
	if os.Getenv("SECRET_KEY") == "" ||
		os.Getenv("ENDPOINT_SECRET") == "" ||
		os.Getenv("SUBSCRIPTION_PRICE_ID") == "" ||
		os.Getenv("DOMAIN") == "" ||
		os.Getenv("LOGS_FILE") == "" ||
		os.Getenv("SECRET_ADMIN") == "" ||
		os.Getenv("NUMBER_OF_SUBSCRIPTIONS_MONTHS") == "" ||
		os.Getenv("STARTING_DATE") == "" {
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
	database.CreateOfflineTable()

	initLogs()

	stripe.Key = os.Getenv("SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir("public"))) //for testing

	http.HandleFunc("/create-checkout-session", createCheckoutSession) //subscricao
	http.HandleFunc("/create-portal-session", createPortalSession)     //para checkar info da subscricao
	http.HandleFunc("/webhook", handleWebhook)

	//testing
	http.HandleFunc("/testeLOGIN", testLogin)

	//auth
	http.HandleFunc("/create-user", createUser)
	http.HandleFunc("/login-user", loginUsr)
	http.HandleFunc("/logout-user", logoutUsr)

	//admin
	http.HandleFunc("/pay-offline", payOffline)
	http.HandleFunc("/get-offline-id", getOfflineWithID)
	http.HandleFunc("/get-offline-last-time", getLastTimeOfflineRequest)

	http.HandleFunc("/health", healthCheck)

	addr := "0.0.0.0:" + port
	log.Printf("Listening on %s", addr)

	// Start HTTPS server
	log.Fatal(http.ListenAndServe(addr, nil))
}
