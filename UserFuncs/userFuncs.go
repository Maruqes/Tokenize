package UserFuncs

import (
	"net/http"
	"strconv"

	"github.com/Maruqes/Tokenize/database"
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
