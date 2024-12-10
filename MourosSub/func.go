package mourosSub

import (
	"os"

	types "github.com/Maruqes/Tokenize/Types"
	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
)

func hadAnySubscription(userID int) (bool, error) {
	usr, err := database.GetUser(userID)
	if err != nil {
		return false, err
	}

	if usr.IsActive {
		return true, nil
	}

	if usr.StripeID == "" {
		return false, nil
	}

	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(usr.StripeID), // Substitua pelo ID do cliente
		Status:   stripe.String("all"),
	}
	i := subscription.List(params)

	for i.Next() {
		s := i.Subscription()
		// Processar cada subscrição conforme necessário
		if s.Items != nil && len(s.Items.Data) > 0 {
			for _, item := range s.Items.Data {
				if item.Price.ID == os.Getenv("SUBSCRIPTION_PRICE_ID") {
					return true, nil
				}
			}
		}
	}
	if err := i.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// this is for mouros subscription, if user had any subscription, return the discount
func returnDisctountStruct(userID int) []*stripe.CheckoutSessionDiscountParams {
	if os.Getenv("COUPON_ID") == "" {
		return nil
	}
	var discounts []*stripe.CheckoutSessionDiscountParams
	val, err := hadAnySubscription(userID)
	if err == nil && val {
		discounts = append(discounts, &stripe.CheckoutSessionDiscountParams{
			Coupon: stripe.String(os.Getenv("COUPON_ID")),
		})
	}
	return discounts
}

// this is for mouros subscription, if user had any subscription, return the discount
func ReturnDisctountStruct(userID int) []*stripe.CheckoutSessionDiscountParams {
	if GLOBAL_TYPE_OF_SUBSCRIPTION != types.TypeOfSubscriptionValues.MourosSubscription {
		var discounts []*stripe.CheckoutSessionDiscountParams
		return discounts
	}

	return returnDisctountStruct(userID)
}
