package checkouts

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Maruqes/Tokenize/UserFuncs"
	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
)

func CheckIfEmailIsBeingUsedInStripe(email string) bool {
	params := &stripe.CustomerListParams{
		Email: stripe.String(email),
	}
	i := customer.List(params)
	for i.Next() {
		fmt.Println(i.Customer().Email)
		if i.Customer().Email == email {
			return true
		}
	}
	return false
}

func CheckIfIDBeingUsedInStripe(id string) bool {
	params := &stripe.CustomerListParams{}
	i := customer.List(params)
	for i.Next() {
		if i.Customer().Metadata["tokenize_id"] == id {
			fmt.Println(i.Customer().ID)
			return true
		}
	}
	return false
}

func GetCustomer(id string) (*stripe.Customer, error) {
	customer, err := customer.Get(id, nil)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, err
	}

	return customer, nil
}

// if custumer already exists in stripe it does not create a new one and uses the existing one
func HandleCreatingCustomer(usr database.User, customer_id string) (*stripe.Customer, error) {
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

		if CheckIfEmailIsBeingUsedInStripe(usr.Email) {
			log.Printf("email already in use")
			return nil, fmt.Errorf("email already in use")
		}

		if CheckIfIDBeingUsedInStripe(customer_id) {
			log.Printf("%s", "id already in use by "+customer_id)
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

func ValidateUserInfoToActivate(customer_id string) (database.User, error) {
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

	//active by offline payment
	endDate, err := UserFuncs.GetEndDateForUser(customerIDInt)
	if err != nil {
		return database.User{}, fmt.Errorf("error getting end date")
	}
	if endDate.Year > time.Now().Year() {
		return database.User{}, fmt.Errorf("user has active offline payment")
	}
	if endDate.Year == time.Now().Year() && endDate.Month > int(time.Now().Month()) {
		return database.User{}, fmt.Errorf("user has active offline payment")
	}
	if endDate.Year == time.Now().Year() && endDate.Month == int(time.Now().Month()) && endDate.Day >= time.Now().Day() {
		return database.User{}, fmt.Errorf("user has active offline payment")
	}

	return usr, nil
}

func GetFixedBillingCycleAnchor(month int, day int) int64 {

	now := time.Now()
	year := now.Year()
	if now.Month() > time.Month(month) || (now.Month() == time.Month(month) && now.Day() > day) {
		year++ // Caso já tenhamos passado a data fixa deste ano, avança para o próximo ano
	}
	fixedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return fixedDate.Unix()
}

func GetFixedBillingFromENV() int64 {
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

	return GetFixedBillingCycleAnchor(monthInt, dayInt)
}
