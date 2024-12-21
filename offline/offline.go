package offline

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Maruqes/Tokenize/database"
)

// does not check if the user is already activated
func ActivateAccountOfflineRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email   string `json:"email"`
		Secret  string `json:"secret"`
		DateEnd string `json:"date_end"` //should be DD/MM/YYYY
	}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if credentials.Email == "" || credentials.Secret == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var date database.Date
	date, err = database.StringToDate(credentials.DateEnd)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	user, err := database.GetUserByEmail(credentials.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	secret_env := os.Getenv("SECRET_ADMIN")

	if credentials.Secret != secret_env {
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return
	}

	todaysDate := database.DateFromUnix(time.Now().Unix())

	err = database.AddOfflinePayment(user.ID, todaysDate, date)
	if err != nil {
		http.Error(w, "Failed to activate account", http.StatusInternalServerError)
		return
	}
}

func IsAccountActivatedOffline(id int) (database.OfflinePayment, error) {
	offlinePayments, err := database.GetOfflinePaymentByID(id)
	if err != nil {
		fmt.Println(err)
		return database.OfflinePayment{}, err
	}

	var lastPayment database.OfflinePayment
	for _, payment := range offlinePayments {
		endDateUnix := time.Date(payment.End_date.Year, time.Month(payment.End_date.Month), payment.End_date.Day, 0, 0, 0, 0, time.UTC).Unix()
		if endDateUnix > time.Now().Unix() {
			lastPayment = payment
		}
	}

	return lastPayment, nil
}

func GetLastEndDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var credentials struct {
		Email  string `json:"email"`
		Secret string `json:"secret"`
	}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if credentials.Email == "" || credentials.Secret == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := database.GetUserByEmail(credentials.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	secret_env := os.Getenv("SECRET_ADMIN")

	if credentials.Secret != secret_env {
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return
	}

	lastPayment, err := IsAccountActivatedOffline(user.ID)
	if err != nil {
		http.Error(w, "Failed to get last payment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(lastPayment.End_date)
}
