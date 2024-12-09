package Tokenize

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	checkouts "github.com/Maruqes/Tokenize/Checkouts"
	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/Logs"
	types "github.com/Maruqes/Tokenize/Types"
	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
)

func handleExtraMouros(userConfirm ExtraPrePayments, db_user database.User) (*stripe.SubscriptionSchedule, error) {
	additionalYears := userConfirm.number_of_payments

	if !functions.CheckMourosDate() || functions.HasStartingDayPassed() {
		additionalYears = additionalYears - 1
	}

	endDate := time.Unix(checkouts.GetFixedBillingFromENV(), 0).AddDate(additionalYears, 0, 0).Unix()

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(db_user.StripeID),
		StartDate: stripe.Int64(time.Now().Unix()),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(int64(1) * int64(userConfirm.number_of_payments)),
					},
				},
				TrialEnd: stripe.Int64(endDate),
				EndDate:  stripe.Int64(endDate),
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

func handleExtraNormal(userConfirm ExtraPrePayments, db_user database.User) (*stripe.SubscriptionSchedule, error) {

	endDate := time.Unix(time.Now().Unix(), 0).AddDate(userConfirm.number_of_payments, 0, 0).Unix()

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(db_user.StripeID),
		StartDate: stripe.Int64(time.Now().Unix()),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(int64(1) * int64(userConfirm.number_of_payments)),
					},
				},
				TrialEnd: stripe.Int64(endDate),
				EndDate:  stripe.Int64(endDate),
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

func handleExtraOnlyOnXNoSubscription(userConfirm ExtraPrePayments, db_user database.User) (*stripe.SubscriptionSchedule, error) {

	endDate := time.Unix(checkouts.GetFixedBillingFromENV(), 0).AddDate(userConfirm.number_of_payments, 0, 0).Unix()

	scheduleParams := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(db_user.StripeID),
		StartDate: stripe.Int64(checkouts.GetFixedBillingFromENV()),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")), // Subscription price ID
						Quantity: stripe.Int64(int64(1) * int64(userConfirm.number_of_payments)),
					},
				},
				TrialEnd: stripe.Int64(endDate),
				EndDate:  stripe.Int64(endDate),
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

func handleExtraPayment(charge stripe.Charge) error {
	purpose := charge.Metadata["purpose"]
	userID := charge.Metadata["user_id"]
	orderID := charge.Metadata["order_id"]
	extraType := charge.Metadata["extra_type"]

	if purpose == "" || userID == "" || orderID == "" || extraType == "" {
		log.Printf("Missing metadata in charge %s", charge.ID)
		return fmt.Errorf("missing metadata in charge %s", charge.ID)
	}

	userConfirm, exists := extra_pagamentos_map[orderID]
	if !exists {
		log.Printf("Order ID %s not found in map", orderID)
		return fmt.Errorf("order ID %s not found in map", orderID)
	}

	if userConfirm.custumerID != userID {
		log.Printf("User not found in map")
		return fmt.Errorf("user not found in map")
	}

	if userConfirm.type_of != "ExtraPayExtra" {
		Logs.PanicLog("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED\n\n")
		fmt.Println("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED")
		return fmt.Errorf("if you seeing this conect support or stop messing with the requests")
	}

	if userConfirm.number_of_payments < 1 {
		log.Printf("Number of payments is less than 1")
		return fmt.Errorf("number of payments is less than 1")
	}

	if extraType != "mbway" && extraType != "multibanco" {
		Logs.PanicLog("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED\n\n")
		fmt.Println("\n\nSHIT IS HAPPENING HERE THIS SHOULD NOT HAPPEN 99% REQUEST ALTERED")
		return fmt.Errorf("if you seeing this conect support or stop messing with the requests")
	}

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

	if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.MourosSubscription {
		schedule, err := handleExtraMouros(userConfirm, db_user)
		if err != nil {
			log.Printf("Error creating subscription schedule: %v", err)
			Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
			return err
		}

		fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)

		log.Println("Payment succeeded for user", userID)
		Logs.LogMessage(fmt.Sprintf("Payment succeeded for user %s", userID))
	} else if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.Normal {
		schedule, err := handleExtraNormal(userConfirm, db_user)
		if err != nil {
			log.Printf("Error creating subscription schedule: %v", err)
			Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
			return err
		}

		fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)

		log.Println("Payment succeeded for user", userID)
		Logs.LogMessage(fmt.Sprintf("Payment succeeded for user %s", userID))
	} else if types.GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.OnlyStartOnDayXNoSubscription {
		schedule, err := handleExtraOnlyOnXNoSubscription(userConfirm, db_user)
		if err != nil {
			log.Printf("Error creating subscription schedule: %v", err)
			Logs.LogMessage(fmt.Sprintf("Error creating subscription schedule: %v", err))
			return err
		}

		fmt.Printf("Subscrição agendada com sucesso! ID: %s\n", schedule.ID)

		log.Println("Payment succeeded for user", userID)
		Logs.LogMessage(fmt.Sprintf("Payment succeeded for user %s", userID))
	}
	return nil
}
