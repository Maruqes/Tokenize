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
		custumer_subscription_deleted(w, event)
	case "customer.subscription.created":
		customer_subscription_created(w, event)
	case "customer.created":
		customer_created(w, event)
	case "invoice.payment_succeeded":
		invoice_payment_succeeded(w, event)
	case "charge.succeeded":
		charge_succeeded(w, event)
	case "invoice.created":
		invoice_created(w, event)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s", event.Type)
		logMessage("Unhandled event type: " + string(event.Type))
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

	http.HandleFunc("/test2", paymentToCreateSubscription)             //subscricao
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
