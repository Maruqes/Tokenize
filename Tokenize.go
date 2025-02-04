package Tokenize

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/Login"
	"github.com/Maruqes/Tokenize/Logs"
	"github.com/Maruqes/Tokenize/StripeFunctions"
	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	portalsession "github.com/stripe/stripe-go/v81/billingportal/session"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/webhook"
)

//falta joia (codigo desconto)

var domain = os.Getenv("DOMAIN")

func createPortalSession(w http.ResponseWriter, r *http.Request) {

	login := Login.CheckToken(r)
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
	ps, err := portalsession.New(params)
	if err != nil {
		log.Printf("Error creating portal session: %v", err)
		http.Error(w, "Failed to create portal session", http.StatusInternalServerError)
		return
	}
	log.Printf("ps.New: %v", ps.URL)
	Logs.LogMessage("Portal session created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)
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
		StripeFunctions.Custumer_subscription_deleted(w, event)
	case "customer.subscription.created":
		StripeFunctions.Customer_subscription_created(w, event)
	case "customer.created":
		StripeFunctions.Customer_created(w, event)
	case "invoice.payment_succeeded":
		StripeFunctions.Invoice_payment_succeeded(w, event)
	case "charge.succeeded":
		StripeFunctions.Charge_succeeded(w, event)
	case "invoice.created":
		StripeFunctions.Invoice_created(w, event)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
		Logs.LogMessage("Unhandled event type: " + string(event.Type))
	}
	w.WriteHeader(http.StatusOK)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	login := Login.CheckToken(r)
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

	log.Printf("Received credentials: %s / %s", credentials.Username, credentials.Email)

	Logs.LogMessage("Create user attempt with email " + credentials.Email + " and username " + credentials.Username)

	if credentials.Username == "" || credentials.Password == "" || credentials.Email == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	isValidEmailTest := functions.IsValidEmail(credentials.Email)
	if !isValidEmailTest {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	//enviar email de confirmacao
	id, err := database.AddUser("", credentials.Email, credentials.Username, credentials.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create user check the credentials with err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	Logs.LogMessage("User created with id/name " + strconv.Itoa(int(id)) + "/" + credentials.Username)

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

	Logs.LogMessage("Login attempt with email " + credentials.Email)

	if credentials.Email == "" || credentials.Password == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	isValidEmailTest := functions.IsValidEmail(credentials.Email)
	if !isValidEmailTest {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	token, usr, err := Login.LoginUser(credentials.Email, credentials.Password)
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
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(5 * 24 * time.Hour),
	})

	Logs.LogMessage("User logged in with id/name " + strconv.Itoa(usr.ID) + "/" + usr.Name)

	w.WriteHeader(http.StatusOK)
}

func logoutUsr(w http.ResponseWriter, r *http.Request) {

	login_Q := Login.CheckToken(r)
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

	Login.LogoutUser(idInt)
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

	Logs.LogMessage("User logged out with id " + strconv.Itoa(idInt))

	w.WriteHeader(http.StatusOK)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

var PRECO_SUB uint64
var lastTimePrecoSub time.Time

func getPrecoSub(w http.ResponseWriter, r *http.Request) {
	if time.Since(lastTimePrecoSub) > 10*time.Minute {
		priceStripe, err := price.Get(os.Getenv("SUBSCRIPTION_PRICE_ID"), nil)
		if err != nil {
			http.Error(w, "Failed to get price", http.StatusInternalServerError)
			return
		}

		PRECO_SUB = uint64(priceStripe.UnitAmount)
		lastTimePrecoSub = time.Now()
	}

	amount := float64(PRECO_SUB) / 100
	response := map[string]float64{"preco": amount}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

}

var initialized bool = false

func Initialize() *sql.DB {
	functions.CheckAllEnv()
	fmt.Println("Init")

	db := database.Init()
	database.CreateTable()
	database.CreatePermissionsTable()

	Logs.InitLogs()

	stripe.Key = os.Getenv("SECRET_KEY")

	initialized = true
	return db
}

func testeEvent(e stripe.Event) {
	fmt.Println("teste")
	metadata, ok := e.Data.Object["metadata"].(map[string]interface{})
	if ok {
		val, ok := metadata["extra"]
		if ok {
			fmt.Println("METADA->:", val)
		}
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	payment, err := StripeFunctions.CreatePaymentPage(1, 49.99, testeEvent, "https://picsum.photos/200/300", "desc", map[string]string{"extra": "testzaomeudeus"})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		return
	}

	//redirect to payment page
	http.Redirect(w, r, payment.URL, http.StatusSeeOther)
}

// set port like "4242"
func InitListen(port string, success string, cancel string) {
	if !initialized {
		Initialize()
	}

	if success[0] != '/' || cancel[0] != '/' {
		panic("Success/Cancel path must start with /")
	}
	port_int, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Invalid port")
	}
	if port_int <= 0 || port_int > 65535 {
		log.Fatal("Invalid port")
	}

	StripeFunctions.PriceID = os.Getenv("SUBSCRIPTION_PRICE_ID")

	http.HandleFunc("/create-portal-session", createPortalSession) //para checkar info da subscricao
	http.HandleFunc("/webhook", handleWebhook)

	//auth
	http.HandleFunc("/create-user", createUser)
	http.HandleFunc("/login-user", loginUsr)
	http.HandleFunc("/logout-user", logoutUsr)

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/getPrecoSub", getPrecoSub)

	// db_user := database.User{1, "", "test0@test0.com", "testemail", false, false}
	// val, err := StripeFunctions.HandleCreatingCustomer(db_user)
	// if err != nil {
	// 	val = val
	// 	log.Fatal("Failed to create customer")
	// }

	if os.Getenv("DEV") == "True" {
		http.Handle("/", http.FileServer(http.Dir("public"))) //for testing
		http.HandleFunc("/test", test)
	}

	addr := "0.0.0.0:" + port
	log.Printf("Listening on %s", addr)

	// Start HTTP server
	log.Fatal(http.ListenAndServe(addr, nil))
}
