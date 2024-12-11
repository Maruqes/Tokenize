package mourosSub

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	checkouts "github.com/Maruqes/Tokenize/Checkouts"
	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/Logs"
	"github.com/Maruqes/Tokenize/database"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
)

func firstSubscriptionMoure(userid string) (*stripe.SubscriptionSchedule, error) {
	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(userid),
		StartDate: stripe.Int64(time.Now().Unix()),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(1),
					},
				},
				TrialEnd: stripe.Int64(checkouts.GetFixedBillingFromENV()),
				EndDate:  stripe.Int64(checkouts.GetFixedBillingFromENV()),
			},
		},
		EndBehavior: stripe.String("cancel"),
	}
	schedule, err := subscriptionschedule.New(scheduleParams)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return nil, err
	}
	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	Logs.LogMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))
	return schedule, nil
}

func secondSubscriptionMoure(userid string) (*stripe.SubscriptionSchedule, error) {

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(userid),
		StartDate: stripe.Int64(checkouts.GetFixedBillingFromENV()), // Future start date in UNIX timestamp
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(1),
					},
				},
				Discounts: returnDisctountStructSchedule(),
			},
		},
	}
	schedule, err := subscriptionschedule.New(scheduleParams)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return nil, err
	}
	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	Logs.LogMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))
	return schedule, nil
}

// Initial Subscription Payment Start Today
func HandleInitialSubscriptionPaymentStartToday(charge stripe.Charge) error {
	purpose := charge.Metadata["purpose"]
	userID := charge.Metadata["user_id"]
	orderID := charge.Metadata["order_id"]

	if purpose == "" || userID == "" || orderID == "" {
		log.Printf("Missing metadata in charge %s", charge.ID)
		return fmt.Errorf("missing metadata in charge %s", charge.ID)
	}

	userConfirm, exists := pagamentos_map[orderID]
	if !exists {
		log.Printf("Order ID %s not found in map", orderID)
		return fmt.Errorf("order ID %s not found in map", orderID)
	}

	if userConfirm.custumerID != userID {
		log.Printf("User not found in map")
		return fmt.Errorf("user not found in map")
	}

	if userConfirm.type_of != "Initial Subscription Payment Start Today" {
		Logs.PanicLog("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED\n\n")
		fmt.Println("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED")
		return fmt.Errorf("if you seeing this conect support or stop messing with the requests")
	}

	log.Println("Payment succeeded for user", userID)
	Logs.LogMessage(fmt.Sprintf("Payment succeeded for user %s", userID))

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Error converting userID to int: %v", err)
		return err
	}
	db_user, err := database.GetUser(userIDInt)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return err
	}

	err = functions.DefinePaymentMethod(db_user.StripeID, charge.PaymentIntent.ID)
	if err != nil {
		log.Printf("Erro ao definir o método de pagamento padrão: %v", err)
		return err
	}

	schedule, err := firstSubscriptionMoure(db_user.StripeID)
	if err != nil {
		log.Printf("Error creating subscription schedule: %v", err)
		Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return err
	}
	fmt.Printf("First Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	Logs.LogMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))

	schedule, err = secondSubscriptionMoure(db_user.StripeID)
	if err != nil {
		log.Printf("Second Error creating subscription schedule: %v", err)
		Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
		return err
	}

	fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)
	Logs.LogMessage(fmt.Sprintf("Subscrição agendada com sucesso! ID: %s", schedule.ID))

	//delete the order from the map
	delete(pagamentos_map, orderID)

	return nil
}
