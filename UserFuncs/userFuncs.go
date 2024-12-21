package UserFuncs

import (
	"fmt"

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
