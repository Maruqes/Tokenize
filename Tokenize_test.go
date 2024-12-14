package Tokenize

import (
	"testing"

	functions "github.com/Maruqes/Tokenize/Functions"
	types "github.com/Maruqes/Tokenize/Types"
	"github.com/Maruqes/Tokenize/database"
)

func TestThis(t *testing.T) {
	Init("4242", "/success.html", "/cancel.html", types.TypeOfSubscriptionValues.MourosSubscription, []types.ExtraPayments{types.ExtraPaymentsValues.Multibanco})
}

func TestPermission7(t *testing.T) {
	database.Init()
	a, err := functions.DoesUserHaveActiveSubscription(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(a)
}
