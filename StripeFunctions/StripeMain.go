package StripeFunctions

import (
	"fmt"
	"time"

	"github.com/Maruqes/Tokenize/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/subscriptionschedule"
)

// PriceID global – define-o com o PriceID correto configurado na tua conta Stripe.
var PriceID string

// Global callback registry: maps callback IDs to functions.
var callbackRegistry = make(map[string]func(event stripe.Event))

// RegisterCallback regista uma callback com um identificador.
func RegisterCallback(id string, cb func(event stripe.Event)) {
	callbackRegistry[id] = cb
}

// getCallback retorna a callback registada para um dado ID.
func getCallback(id string) func(event stripe.Event) {
	return callbackRegistry[id]
}

// generateCallbackID gera um ID único para a callback.
func generateCallbackID() string {
	return fmt.Sprintf("cb_%d", time.Now().UnixNano())
}

// --- Funções de criação de objetos Stripe ---

// CreateSubscription cria uma subscrição normal com um período de trial (ex: 1 ano).
// Regista a callback e define explicitamente os metadata (incluindo o callbackID).
func CreateSubscription(userID string, duration time.Duration, callback func(event stripe.Event)) (*stripe.Subscription, error) {
	trialEnd := time.Now().Add(duration).Unix()

	callbackID := generateCallbackID()
	RegisterCallback(callbackID, callback)

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(userID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(PriceID),
			},
		},
		TrialEnd: stripe.Int64(trialEnd),
		Metadata: map[string]string{
			"callback": callbackID,
		},
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// CreateScheduledSubscription cria uma subscrição agendada que inicia na data 'start'
// e tem uma duração definida (ex: 1 ano). Regista a callback e define explicitamente os metadata.
func CreateScheduledSubscription(userID string, start time.Time, duration time.Duration, callback func(event stripe.Event)) (*stripe.SubscriptionSchedule, error) {
	phaseEnd := start.Add(duration).Unix()

	callbackID := generateCallbackID()
	RegisterCallback(callbackID, callback)

	params := &stripe.SubscriptionScheduleParams{
		Customer:    stripe.String(userID),
		StartDate:   stripe.Int64(start.Unix()),
		EndBehavior: stripe.String("cancel"),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{
						Price: stripe.String(PriceID),
					},
				},
				EndDate: stripe.Int64(phaseEnd),
			},
		},
		Metadata: map[string]string{
			"callback": callbackID,
		},
	}

	schedule, err := subscriptionschedule.New(params)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

// CreateFreeTrial cria uma subscrição com período de trial gratuito.
// O trial inicia em 'start' e termina após a duração indicada.
// Regista a callback e define explicitamente os metadata.
func CreateFreeTrial(userID string, start time.Time, duration time.Duration, callback func(event stripe.Event)) (*stripe.Subscription, error) {
	trialEnd := start.Add(duration).Unix()

	callbackID := generateCallbackID()
	RegisterCallback(callbackID, callback)

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(userID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(PriceID),
			},
		},
		TrialEnd: stripe.Int64(trialEnd),
		Metadata: map[string]string{
			"callback": callbackID,
		},
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// CreatePayment cria um PaymentIntent para efetuar um pagamento.
// O valor 'amount' é fornecido em unidades (ex: 49.99) e convertido para cêntimos.
// Regista a callback e define explicitamente os metadata.
func CreatePayment(userID int, amount float64, callback func(event stripe.Event)) (*stripe.PaymentIntent, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	amt := int64(amount * 100)

	callbackID := generateCallbackID()
	RegisterCallback(callbackID, callback)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amt),
		Currency: stripe.String(string(stripe.CurrencyEUR)),
		Customer: stripe.String(usrDB.StripeID),
		Metadata: map[string]string{
			"callback": callbackID,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// CreatePaymentPage cria uma sessão de Checkout da Stripe para pagamento.
// Permite definir imagem, descrição e metadados extra para personalizar a página.
// Regista a callback e define explicitamente os metadata tanto para a sessão quanto para o PaymentIntent.
func CreatePaymentPage(userID int, amount float64, callback func(event stripe.Event), imageURL, description string, extraMetadata map[string]string) (*stripe.CheckoutSession, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	amt := int64(amount * 100)

	callbackID := generateCallbackID()
	RegisterCallback(callbackID, callback)

	// Cria o mapa de metadata com o callbackID e junta os extraMetadata (se houver)
	metadata := map[string]string{
		"callback": callbackID,
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	// Define os parâmetros conforme o snippet fornecido:
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("payment"),
		Customer:           stripe.String(usrDB.StripeID),
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
		SuccessURL: stripe.String("https://localhost/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("https://localhost/cancel"),
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

// CreateSubscriptionPage cria uma sessão de Checkout da Stripe para subscrição,
// utilizando o PriceID indicado. Permite definir imagem, descrição e metadados extra.
// Regista a callback e define explicitamente os metadata tanto para a sessão quanto para os dados da subscrição.
func CreateSubscriptionPage(userID int, priceID string, callback func(event stripe.Event), imageURL, description string, extraMetadata map[string]string) (*stripe.CheckoutSession, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	callbackID := generateCallbackID()
	RegisterCallback(callbackID, callback)

	metadata := map[string]string{
		"callback": callbackID,
	}
	for key, value := range extraMetadata {
		metadata[key] = value
	}

	// Para Checkout Session em modo de subscrição usamos SubscriptionData em vez de PaymentIntentData.
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("subscription"),
		Customer:           stripe.String(usrDB.StripeID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://localhost/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("https://localhost/cancel"),
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

// CheckUserPaymentMethod verifica se o cliente (identificado pelo userID na base de dados)
// tem registado um método de pagamento (por exemplo, cartão) na Stripe.
func CheckUserPaymentMethod(userID int) (bool, error) {
	usrDB, err := database.GetUser(userID)
	if err != nil {
		return false, err
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
