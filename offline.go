package Tokenize

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/Maruqes/Tokenize/database"
)

func DailyCheckRoutine(targetHour int, targetMinute int) {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMinute, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		// Sleep until the next run time.
		sleepDuration := time.Until(nextRun)
		log.Printf("Daily task scheduled to run in: %v\n", sleepDuration)
		time.Sleep(sleepDuration)

		// Perform the daily task.
		log.Println("Checking for offline payments...")
	}
}

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

	login_Q := checkToken(r)
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

	login_Q := checkToken(r)
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

func getLastTimeOffline(userID int) (database.Date, error) {
	offlinePayments, err := database.GetOfflinePaymentByID(userID)
	if err != nil {
		return database.Date{}, err
	}

	sort.Slice(offlinePayments, func(i, j int) bool {
		dateI := time.Date(offlinePayments[i].DateOfPayment.Year, time.Month(offlinePayments[i].DateOfPayment.Month), offlinePayments[i].DateOfPayment.Day, 0, 0, 0, 0, time.UTC)
		dateJ := time.Date(offlinePayments[j].DateOfPayment.Year, time.Month(offlinePayments[j].DateOfPayment.Month), offlinePayments[j].DateOfPayment.Day, 0, 0, 0, 0, time.UTC)
		return dateI.Before(dateJ)
	})

	numberOfSubscriptionDays, err := strconv.Atoi(os.Getenv("NUMBER_OF_SUBSCRIPTIONS_DAYS"))
	if err != nil {
		return database.Date{}, fmt.Errorf("invalid NUMBER_OF_SUBSCRIPTIONS_DAYS: %v", err)
	}
	time_each_sub := (24 * time.Hour) * time.Duration(numberOfSubscriptionDays)
	last_time := time.Now()
	set_last := false

	for i := 0; i < len(offlinePayments); i++ {
		date := time.Date(offlinePayments[i].DateOfPayment.Year, time.Month(offlinePayments[i].DateOfPayment.Month), offlinePayments[i].DateOfPayment.Day, 0, 0, 0, 0, time.UTC)
		date_end := date.Add(time_each_sub * time.Duration(offlinePayments[i].Quantity))
		fmt.Println(date)
		fmt.Println(date_end)

		if !set_last {
			if date_end.After(last_time) {
				last_time = date_end
				set_last = true
				fmt.Println("to_compare")
			}
		} else {
			last_time = last_time.Add(time_each_sub * time.Duration(offlinePayments[i].Quantity))
		}

		fmt.Printf("\n\n\n")
	}

	fmt.Println(last_time)

	return offlinePayments[len(offlinePayments)-1].DateOfPayment, nil
}

func getLastTimeOfflineRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login_Q := checkToken(r)
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
