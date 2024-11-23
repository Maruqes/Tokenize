package Tokenize

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/Maruqes/Tokenize/database"
)

func validateDate(date database.Date) bool {
	if date.Day < 1 || date.Day > 31 {
		return false
	}

	if date.Month < 1 || date.Month > 12 {
		return false
	}

	if date.Year < 1069 || date.Year > 9999 {
		return false
	}

	return true
}

func payOffline(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login_Q := CheckToken(r)
	if !login_Q {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	var data struct {
		SecretAdmin string `json:"secret_admin"`
		UserID      string `json:"user_id"`
		Quantity    int    `json:"quantity"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if data.SecretAdmin != os.Getenv("SECRET_ADMIN") {
		http.Error(w, "Invalid auth", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(data.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if data.Quantity <= 0 {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
	}

	if exists, err := database.CheckIfUserIDExists(userID); err != nil || !exists {
		http.Error(w, "User does not exist", http.StatusBadRequest)
		return
	}

	//user com subscription ativa por stripe nao pode pagar offline
	usr, err := database.GetUser(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if usr.IsActive {
		http.Error(w, "User already has an active subscription", http.StatusBadRequest)
		return
	}

	date := time.Now()
	dateOfPayment := database.Date{
		Day:   date.Day(),
		Month: int(date.Month()),
		Year:  date.Year(),
	}

	if !validateDate(dateOfPayment) {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	err = database.AddOfflinePayment(userID, dateOfPayment, data.Quantity)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Failed to retrieve cookie", http.StatusInternalServerError)
		return
	}
	logMessage(fmt.Sprintf("User %d paid %d offline, AUTHORIZER_ID: %s", userID, data.Quantity, cookie.Value))

	w.WriteHeader(http.StatusOK)
}

func getOfflineWithID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login_Q := CheckToken(r)
	if !login_Q {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}
	var data struct {
		SecretAdmin string `json:"secret_admin"`
		UserID      string `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if data.SecretAdmin != os.Getenv("SECRET_ADMIN") {
		http.Error(w, "Invalid auth", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(data.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	offlinePayments, err := database.GetOfflinePaymentByID(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(offlinePayments)
}

func getLastTimeAlgorithm(offlinePayments []database.OfflinePayment) (time.Time, error) {
	sort.Slice(offlinePayments, func(i, j int) bool {
		dateI := time.Date(offlinePayments[i].DateOfPayment.Year, time.Month(offlinePayments[i].DateOfPayment.Month), offlinePayments[i].DateOfPayment.Day, 0, 0, 0, 0, time.UTC)
		dateJ := time.Date(offlinePayments[j].DateOfPayment.Year, time.Month(offlinePayments[j].DateOfPayment.Month), offlinePayments[j].DateOfPayment.Day, 0, 0, 0, 0, time.UTC)
		return dateI.Before(dateJ)
	})

	numberOfSubscriptionMonths, err := strconv.Atoi(os.Getenv("NUMBER_OF_SUBSCRIPTIONS_MONTHS"))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid NUMBER_OF_SUBSCRIPTIONS_MONTHS: %v", err)
	}

	var expiryDate time.Time

	for _, payment := range offlinePayments {
		paymentDate := time.Date(payment.DateOfPayment.Year, time.Month(payment.DateOfPayment.Month), payment.DateOfPayment.Day, 0, 0, 0, 0, time.UTC)
		addedTime := numberOfSubscriptionMonths * payment.Quantity

		if paymentDate.After(expiryDate) {
			// If the payment is after the current expiry date, start a new subscription
			expiryDate = paymentDate.AddDate(0, addedTime, 0)
		} else {
			// Extend the current subscription
			expiryDate = expiryDate.AddDate(0, addedTime, 0)
		}
	}

	return expiryDate, nil
}

func getLastTimeOffline(userID int) (database.Date, error) {
	if ol, err := database.DoesUserHasOfflinePayments(userID); !ol || err != nil {
		return database.Date{}, nil
	}

	offlinePayments, err := database.GetOfflinePaymentByID(userID)
	if err != nil {
		return database.Date{}, err
	}

	last_time, err := getLastTimeAlgorithm(offlinePayments)
	if err != nil {
		return database.Date{}, err
	}

	return database.Date{
		Day:   last_time.Day(),
		Month: int(last_time.Month()),
		Year:  last_time.Year(),
	}, nil
}

func getLastTimeOfflineRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login_Q := CheckToken(r)
	if !login_Q {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}
	var data struct {
		SecretAdmin string `json:"secret_admin"`
		UserID      string `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if data.SecretAdmin != os.Getenv("SECRET_ADMIN") {
		http.Error(w, "Invalid auth", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(data.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	lastTime, err := getLastTimeOffline(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(lastTime)
}

func doesHaveOfflinePayments(userID int) (bool, error) {
	lastTime, err := getLastTimeOffline(userID)
	if err != nil {
		return false, err
	}

	if (lastTime == database.Date{}) {
		return false, nil
	}

	if time.Now().After(time.Date(lastTime.Year, time.Month(lastTime.Month), lastTime.Day, 0, 0, 0, 0, time.UTC)) {
		return false, nil
	}

	return true, nil
}
