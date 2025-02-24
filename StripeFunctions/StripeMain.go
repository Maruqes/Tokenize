package StripeFunctions

import (
	"time"

	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
)

func CreateSubscription(userID int, trial_duration time.Duration, PriceID string, extraMetadata map[string]string) (*stripe.Subscription, error) {
	trialEnd := time.Now().Add(trial_duration).Unix()

	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	usrStripe, err := HandleCreatingCustomer(usrDB)
	if err != nil {
		return nil, err
	}

	// Cria o mapa de metadata com o callbackID e junta os extraMetadata (se houver)
	metadata := map[string]string{
		"callback": "CreateSubscription",
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(usrStripe.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(PriceID),
			},
		},
		TrialEnd: stripe.Int64(trialEnd),
		Metadata: metadata,
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// callback only calls when the start time is reached
func CreateScheduledSubscription(userID int, start time.Time, trial_duration time.Duration, PriceID string, extraMetadata map[string]string) (*stripe.SubscriptionSchedule, error) {
	trialEnd := start.Add(trial_duration).Unix()

	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	usrStripe, err := HandleCreatingCustomer(usrDB)
	if err != nil {
		return nil, err
	}

	// Cria o mapa de metadata com o callbackID e junta os extraMetadata (se houver)
	metadata := map[string]string{
		"callback": "CreateScheduledSubscription",
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	params := &stripe.SubscriptionScheduleParams{
		Customer:  stripe.String(usrStripe.ID),
		StartDate: stripe.Int64(start.Unix()),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price: stripe.String(PriceID),
					},
				},
				TrialEnd: stripe.Int64(trialEnd),
			},
		},
		Metadata: metadata,
	}

	schedule, err := subscriptionschedule.New(params)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func CreateFreeTrial(userID int, start time.Time, duration time.Duration, PriceID string, extraMetadata map[string]string) (*stripe.Subscription, error) {
	trialEnd := start.Add(duration).Unix()

	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	usrStripe, err := HandleCreatingCustomer(usrDB)
	if err != nil {
		return nil, err
	}

	// Cria o mapa de metadata com o callbackID e junta os extraMetadata (se houver)
	metadata := map[string]string{
		"callback": "CreateFreeTrial",
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	cancelAt := trialEnd - (86400 / 24) // Cancel subscription 1 hours before trial end
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(usrStripe.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(PriceID),
			},
		},
		TrialEnd: stripe.Int64(trialEnd),
		CancelAt: stripe.Int64(cancelAt),
		Metadata: metadata,
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func CreatePayment(userID int, amount float64, extraMetadata map[string]string) (*stripe.PaymentIntent, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	usrStripe, err := HandleCreatingCustomer(usrDB)
	if err != nil {
		return nil, err
	}

	amt := int64(amount * 100)

	// Cria o mapa de metadata com o callbackID e junta os extraMetadata (se houver)
	metadata := map[string]string{
		"callback": "CreatePayment",
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amt),
		Currency: stripe.String(string(stripe.CurrencyEUR)),
		Customer: stripe.String(usrStripe.ID),
		Metadata: metadata,
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

func CreatePaymentPage(userID int, amount float64, imageURL, description string,
	extraMetadata map[string]string, success_url string, cancel_url string) (*stripe.CheckoutSession, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	usrStripe, err := HandleCreatingCustomer(usrDB)
	if err != nil {
		return nil, err
	}

	amt := int64(amount * 100)

	// Cria o mapa de metadata com o callbackID e junta os extraMetadata (se houver)
	metadata := map[string]string{
		"callback": "CreatePaymentPage",
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("payment"),
		Customer:           stripe.String(usrStripe.ID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("eur"),
					UnitAmount: stripe.Int64(amt),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Payment for service"),
						Description: stripe.String(description),
						Images:      stripe.StringSlice([]string{imageURL}),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(success_url),
		CancelURL:  stripe.String(cancel_url),
		Metadata:   metadata,
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metadata,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func CreateSubscriptionPage(userID int, priceID string, extraMetadata map[string]string,
	success_url string, cancel_url string) (*stripe.CheckoutSession, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	usrStripe, err := HandleCreatingCustomer(usrDB)
	if err != nil {
		return nil, err
	}

	metadata := map[string]string{
		"callback": "CreateSubscriptionPage",
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	// Para Checkout Session em modo de subscrição usamos SubscriptionData em vez de PaymentIntentData.
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("subscription"),
		Customer:           stripe.String(usrStripe.ID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(success_url),
		CancelURL:  stripe.String(cancel_url),
		Metadata:   metadata,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func CheckUserPaymentMethod(userID int) (bool, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return false, err
	}
	if usrDB.StripeID == "" {
		return false, nil
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(usrDB.StripeID),
		Type:     stripe.String("card"),
	}

	i := paymentmethod.List(params)
	for i.Next() {
		pm := i.PaymentMethod()
		if pm != nil {
			return true, nil
		}
	}
	if err := i.Err(); err != nil {
		return false, err
	}
	return false, nil
}
