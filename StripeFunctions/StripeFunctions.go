package StripeFunctions

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
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
func HandleCreatingCustomer(usr database.User) (*stripe.Customer, error) {

	if usr.Email == "" {
		fmt.Println("user email is empty")
		return nil, fmt.Errorf("user email is empty")
	}
	customer_id := strconv.Itoa(usr.ID)

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

func GetEndDateUserStripe(userId int) (database.Date, error) {
	user, err := database.GetUser(userId)
	if err != nil {
		return database.Date{}, err
	}

	if user.StripeID == "" {
		return database.Date{}, fmt.Errorf("no end date available or no stripe id")
	}

	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(user.StripeID),
		Status:   stripe.String("all"), // Include all statuses to catch trials
	}

	var lastEnd int64
	lastEnd = 0

	i := subscription.List(params)
	for i.Next() {
		s := i.Subscription()

		// Consider the greater of CurrentPeriodEnd, CancelAt, and TrialEnd
		subEnd := s.CurrentPeriodEnd
		if s.TrialEnd > subEnd {
			subEnd = s.TrialEnd
		}
		if s.CancelAt > 0 {
			continue //if canceled, don't consider
		}
		if subEnd > lastEnd {
			lastEnd = subEnd
		}
	}

	if lastEnd == 0 {
		return database.Date{}, fmt.Errorf("no end date available or no stripe id")
	}

	return database.DateFromUnix(lastEnd), nil
}

// subscription funs
type Subscription struct {
	ID         string        `json:"id"`
	ScheduleID string        `json:"ScheduleID"`
	UserID     int           `json:"user_id"`
	StartDate  database.Date `json:"start_date"`
	EndDate    database.Date `json:"end_date"`
	Active     bool          `json:"active"`
	Trial      bool          `json:"trial"`
	Used       bool          `json:"used"`
	Schedule   bool          `json:"schedule"`
}

func (s Subscription) String() string {
	return fmt.Sprintf("ID: %s\nSheduleID: %s\nUserID: %d\nStartDate: %s\nEndDate: %s\nActive: %t\nTrial: %t\nUsed: %t\nSchedule: %t\n",
		s.ID,
		s.ScheduleID,
		s.UserID,
		s.StartDate.String(),
		s.EndDate.String(),
		s.Active,
		s.Trial,
		s.Used,
		s.Schedule)
}

func getNormalSubs(user database.User, wg *sync.WaitGroup, res *[]Subscription) {
	defer wg.Done()
	// Fetch active subscriptions
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(user.StripeID),
		Status:   stripe.String("all"),
	}

	i := subscription.List(params)
	for i.Next() {
		s := i.Subscription()
		startDate := time.Unix(s.CurrentPeriodStart, 0)
		endDate := time.Unix(s.CurrentPeriodEnd, 0)
		currentDate := time.Now()

		used := currentDate.After(startDate) && currentDate.Before(endDate.Add(24*time.Hour))

		subsS := Subscription{
			ID:         s.ID,
			ScheduleID: "",
			UserID:     user.ID,
			StartDate:  database.DateFromUnix(s.CurrentPeriodStart),
			EndDate:    database.DateFromUnix(s.CurrentPeriodEnd),
			Active:     s.Status == "active",
			Trial:      s.Status == "trialing",
			Used:       used,
			Schedule:   false,
		}

		*res = append(*res, subsS)
	}
}

func getScheduledSubs(user database.User, wg *sync.WaitGroup, res *[]Subscription) {
	defer wg.Done()
	// Fetch active subscriptions
	scheduleParams := &stripe.SubscriptionScheduleListParams{
		Customer: stripe.String(user.StripeID),
	}

	scheduleList := subscriptionschedule.List(scheduleParams)
	for scheduleList.Next() {
		schedule := scheduleList.SubscriptionSchedule()

		for _, phase := range schedule.Phases {

			subsS := Subscription{
				ID:         "",
				ScheduleID: schedule.ID,
				UserID:     user.ID,
				StartDate:  database.DateFromUnix(phase.StartDate),
				EndDate:    database.DateFromUnix(phase.EndDate),
				Active:     schedule.Status == "active",
				Trial:      schedule.Status == "trialing",
				Used:       false,
				Schedule:   true,
			}

			// Verifica se o ID da subscrição não é nulo ou vazio
			if schedule.Subscription != nil && schedule.Subscription.ID != "" {
				subsS.ID = schedule.Subscription.ID
			}

			*res = append(*res, subsS)
		}
	}
}

func GetAllSubscriptions(userID int) ([]Subscription, error) {
	var wg sync.WaitGroup
	var res []Subscription

	user, err := database.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user")
	}

	wg.Add(2)

	go getNormalSubs(user, &wg, &res)
	go getScheduledSubs(user, &wg, &res)

	wg.Wait()

	return res, nil
}

func GetUserIdWithStripeID(stripeID string) (int, error) {
	user, err := database.GetUserByStripeID(stripeID)
	if err != nil {
		return -1, err
	}
	return user.ID, nil
}
