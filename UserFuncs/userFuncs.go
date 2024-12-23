package UserFuncs

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/database"
	"github.com/Maruqes/Tokenize/offline"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
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
	lastDateOffline, err := offline.GetLastEndDate(id)
	if err != nil {
		return database.Date{}, fmt.Errorf("error catching offline payment")
	}

	lastStripePayment, err := functions.GetEndDateUserStripe(id)
	if err != nil {
		if err.Error() != "no end date available" {
			return database.Date{}, fmt.Errorf("error catching stripe payment")
		}
	}

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

// subscription funs
type Subscription struct {
	ID        string        `json:"id"`
	UserID    int           `json:"user_id"`
	StartDate database.Date `json:"start_date"`
	EndDate   database.Date `json:"end_date"`
	Active    bool          `json:"active"`
	Trial     bool          `json:"trial"`
	Used      bool          `json:"used"`
	Schedule  bool          `json:"schedule"`
}

type SubscriptionInterface interface {
	Print() string
}

func (s Subscription) String() string {
	return fmt.Sprintf("ID: %s\nUserID: %d\nStartDate: %s\nEndDate: %s\nActive: %t\nTrial: %t\nUsed: %t\nSchedule: %t\n",
		s.ID,
		s.UserID,
		s.StartDate.String(),
		s.EndDate.String(),
		s.Active,
		s.Trial,
		s.Used,
		s.Schedule)
}

func isSubscriptionInResult(sub *stripe.Subscription, res []Subscription) bool {
	if sub == nil {
		return false
	}

	for _, s := range res {
		if s.ID == sub.ID {
			return true
		}
	}

	return false
}

func GetAllSubscriptions(userID int) ([]Subscription, error) {
	var res []Subscription

	user, err := database.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user")
	}

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
			ID:        s.ID,
			UserID:    userID,
			StartDate: database.DateFromUnix(s.CurrentPeriodStart),
			EndDate:   database.DateFromUnix(s.CurrentPeriodEnd),
			Active:    s.Status == "active",
			Trial:     s.Status == "trialing",
			Used:      used,
			Schedule:  false,
		}

		res = append(res, subsS)
	}

	// Fetch scheduled subscriptions
	scheduleParams := &stripe.SubscriptionScheduleListParams{
		Customer: stripe.String(user.StripeID),
	}

	scheduleList := subscriptionschedule.List(scheduleParams)
	for scheduleList.Next() {
		schedule := scheduleList.SubscriptionSchedule()

		for _, phase := range schedule.Phases {

			if isSubscriptionInResult(schedule.Subscription, res) {
				continue
			}

			subsS := Subscription{
				ID:        schedule.ID,
				UserID:    userID,
				StartDate: database.DateFromUnix(phase.StartDate),
				EndDate:   database.DateFromUnix(phase.EndDate),
				Active:    schedule.Status == "active",
				Trial:     schedule.Status == "trialing",
				Used:      false,
				Schedule:  true,
			}
			res = append(res, subsS)
		}
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("error fetching subscriptions: %v", err)
	}

	return res, nil
}
