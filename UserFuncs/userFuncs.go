package UserFuncs

import (
	"fmt"
	"net/http"
	"strconv"

	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/database"
	"github.com/Maruqes/Tokenize/offline"
)

func GetAllUsers() ([]database.User, error) {
	return database.GetAllUsers()
}

func GetUserByID(id int) (database.User, error) {
	return database.GetUser(id)
}

func GetUserByEmail(email string) (database.User, error) {
	return database.GetUserByEmail(email)
}

func GetEndDateForUser(id int) (database.Date, error) {
	lastDateOffline, err := offline.IsAccountActivatedOffline(id)
	if err != nil {
		return database.Date{}, fmt.Errorf("error catching offline payment")
	}

	lastStripePayment, err := functions.GetEndDateUserStripe(id)
	if err != nil {
		return database.Date{}, fmt.Errorf("error catching stripe payment")
	}

	fmt.Printf("lastDateOffline: %v\n", lastDateOffline)
	fmt.Printf("lastStripePayment: %v\n", lastStripePayment)

	if lastDateOffline.End_date.Year > lastStripePayment.Year {
		return lastDateOffline.End_date, nil
	} else if lastDateOffline.End_date.Year < lastStripePayment.Year {
		return lastStripePayment, nil
	}

	if lastDateOffline.End_date.Month > lastStripePayment.Month {
		return lastDateOffline.End_date, nil
	} else if lastDateOffline.End_date.Month < lastStripePayment.Month {
		return lastStripePayment, nil
	}

	if lastDateOffline.End_date.Day > lastStripePayment.Day {
		return lastDateOffline.End_date, nil
	} else if lastDateOffline.End_date.Day < lastStripePayment.Day {
		return lastStripePayment, nil
	}

	if lastDateOffline.End_date.Day == lastStripePayment.Day {
		return lastDateOffline.End_date, nil
	}

	return database.Date{}, fmt.Errorf("error catching end date")
}

func ProhibitUser(id int) error {
	return database.ProhibitUser(id)
}

func UnprohibitUser(id int) error {
	return database.UnprohibitUser(id)
}

func IsProhibited(id int) (bool, error) {
	return database.CheckIfUserIsProhibited(id)
}

// assumes that the user is already validated
func CheckProhibitedUser(w http.ResponseWriter, r *http.Request) bool {
	//get id
	customer_id_cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Error getting id", http.StatusInternalServerError)
		return false
	}

	customer_id := customer_id_cookie.Value
	customerIDInt, err := strconv.Atoi(customer_id)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return false
	}

	prohibited, err := IsProhibited(customerIDInt)
	if err != nil {
		http.Error(w, "Error checking if user is prohibited", http.StatusInternalServerError)
		return false
	}

	if prohibited {
		http.Error(w, "User is prohibited", http.StatusForbidden)
		return true
	}

	return false
}

func ActivateUser(id int) error {
	return database.ActivateUser(id)
}

func DeactivateUser(id int) error {
	return database.DeactivateUser(id)
}
