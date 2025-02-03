package Tokenize

import (
	"testing"

	types "github.com/Maruqes/Tokenize/Types"
)

func TestThis(t *testing.T) {
	Initialize()
	InitListen("4242", "/success.html", "/cancel.html", types.TypeOfSubscriptionValues.MourosSubscription, []types.ExtraPayments{types.ExtraPaymentsValues.Multibanco})
}
