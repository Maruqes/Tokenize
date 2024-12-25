package offline

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	funchooks "github.com/Maruqes/Tokenize/FuncHooks"
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

	//verify secret
	if credentials.Secret != secret_env {
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return
	}

	if funchooks.PayOffline_UserFunc != nil {
		if funchooks.PayOffline_UserFunc(w, r) {
			return
		}
	}

	todaysDate := database.DateFromUnix(time.Now().Unix())

	err = database.AddOfflinePayment(user.ID, todaysDate, date)
	if err != nil {
		http.Error(w, "Failed to activate account", http.StatusInternalServerError)
		return
	}
}

func GetLastEndDate(id int) (database.OfflinePayment, error) {
	all_offline, err := database.GetOfflinePaymentByID(id)
	if err != nil {
		return database.OfflinePayment{}, err
	}

	if len(all_offline) == 0 {
		return database.OfflinePayment{}, nil
	}

	var last_offline database.OfflinePayment
	last_offline = all_offline[0]

	for _, off := range all_offline {
		if off.End_date.Year > last_offline.End_date.Year {
			last_offline = off
		} else if off.End_date.Year == last_offline.End_date.Year {
			if off.End_date.Month > last_offline.End_date.Month {
				last_offline = off
			} else if off.End_date.Month == last_offline.End_date.Month {
				if off.End_date.Day > last_offline.End_date.Day {
					last_offline = off
				}
			}
		}
	}

	return last_offline, nil
}
