package normalSub

import (
	"log"
	"net/http"
	"os"

	checkouts "github.com/Maruqes/Tokenize/Checkouts"
	"github.com/Maruqes/Tokenize/Login"
	"github.com/Maruqes/Tokenize/Logs"
	types "github.com/Maruqes/Tokenize/Types"
	"github.com/Maruqes/Tokenize/UserFuncs"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

/*
card, acss_debit, affirm, afterpay_clearpay, alipay, au_becs_debit,
bacs_debit, bancontact, blik, boleto, cashapp, customer_balance, eps,
fpx, giropay, grabpay, ideal, klarna, konbini, link, multibanco, oxxo,
p24, paynow, paypal, pix, promptpay, sepa_debit, sofort, swish, us_bank_account,
wechat_pay, revolut_pay, mobilepay, zip, amazon_pay, alma, twint, kr_card,
naver_pay, kakao_pay, payco, or samsung_pay"
*/

var (
	domain                      string
	success_path                string
	cancel_path                 string
	GLOBAL_TYPE_OF_SUBSCRIPTION types.TypeOfSubscription
)

func InitNormalCheckouts(d string, sp string, cp string, gtos types.TypeOfSubscription) {
	domain = d
	success_path = sp
	cancel_path = cp
	GLOBAL_TYPE_OF_SUBSCRIPTION = gtos
}

func createCheckoutStruct(finalCustomer *stripe.Customer) *stripe.CheckoutSessionParams {
	if GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.OnlyStartOnDayXNoSubscription {
		panic("This function should not be called with this type of subscription")
	}

	if GLOBAL_TYPE_OF_SUBSCRIPTION == types.TypeOfSubscriptionValues.OnlyStartOnDayX {
		return &stripe.CheckoutSessionParams{
			Customer: stripe.String(finalCustomer.ID),
			Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")),
					Quantity: stripe.Int64(1),
				},
			},
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				BillingCycleAnchor: stripe.Int64(checkouts.GetFixedBillingFromENV()), // Função personalizada para calcular o timestamp
			},

			PaymentMethodTypes: stripe.StringSlice([]string{
				"card", // Standard credit/debit card payment method
			}),
			SuccessURL: stripe.String(domain + success_path),
			CancelURL:  stripe.String(domain + cancel_path),
		}
	}

	return &stripe.CheckoutSessionParams{
		Customer: stripe.String(finalCustomer.ID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(os.Getenv("SUBSCRIPTION_PRICE_ID")),
				Quantity: stripe.Int64(1),
			},
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card", // Standard credit/debit card payment method
		}),

		SuccessURL: stripe.String(domain + success_path),
		CancelURL:  stripe.String(domain + cancel_path),
	}
}

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	login := Login.CheckToken(r)
	if !login {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	// Check if user is prohibited and respond accordingly
	prohibited := UserFuncs.CheckProhibitedUser(w, r)
	if prohibited {
		return
	}

	//get id
	customer_id_cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, "Error getting id", http.StatusInternalServerError)
		return
	}
	customer_id := customer_id_cookie.Value

	//validate user
	usr, err := checkouts.ValidateUserInfoToActivate(customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//creates or gets the customer
	finalCustomer, err := checkouts.HandleCreatingCustomer(usr, customer_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Configurar sessão de checkout com o cliente criado
	checkoutParams := createCheckoutStruct(finalCustomer)

	if checkouts.GetFixedBillingFromENV() == 0 {
		checkoutParams.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			BillingCycleAnchor: nil,
		}
	}

	s, err := session.New(checkoutParams)
	if err != nil {
		log.Printf("session.New: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	Logs.LogMessage("Checkout session created for user " + usr.Name + " with id " + customer_id + " and email " + usr.Email)

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}
