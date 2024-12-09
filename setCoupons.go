package Tokenize

import (
	"fmt"
	"log"
	"os"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
)

func SetCoupon(subscriptionStripeID string, couponID ...string) {
	var couponID_var string
	if len(couponID) == 0 {
		couponID_var = os.Getenv("COUPON_ID")
		if couponID_var == "" {
			log.Println("No coupon ID provided")
			return
		}
	} else {
		couponID_var = couponID[0]
	}

	// Atualizar a subscrição para adicionar o cupão
	params := &stripe.SubscriptionParams{
		Coupon: stripe.String(couponID_var),
	}

	updatedSubscription, err := subscription.Update(subscriptionStripeID, params)
	if err != nil {
		log.Fatalf("Erro ao atualizar subscrição: %v", err)
	}

	// Confirmar se o cupão foi aplicado com sucesso
	fmt.Printf("Subscrição atualizada com sucesso: %+v\n", updatedSubscription.ID)

}
