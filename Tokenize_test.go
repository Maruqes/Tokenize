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

var perms = permissions{}

func TestPermission1(t *testing.T) {
	database.Init()
	err := perms.CreatePermission("balcao", "bar:write")
	if err != nil {
		t.Error(err)
	}
}

func TestPermission2(t *testing.T) {
	database.Init()
	err := perms.DeletePermission(5)
	if err != nil {
		t.Error(err)
	}
}

func TestPermission3(t *testing.T) {
	database.Init()
	err := perms.AddUserPermission(1, 1)
	if err != nil {
		t.Error(err)
	}
}

func TestPermission4(t *testing.T) {
	database.Init()
	err := perms.RemoveUserPermission(1, 1)
	if err != nil {
		t.Error(err)
	}

}

func TestPermission5(t *testing.T) {
	database.Init()
	perms, err := perms.GetUserPermissions(1)
	t.Log(perms)

	if err != nil {
		t.Error(err)
	}
	if len(perms) == 0 {
		t.Error("no permissions")
	}
	t.Log("Permissions len: ", len(perms))
	for _, perm := range perms {
		t.Log(perm)
	}
}

func TestPermission6(t *testing.T) {
	database.Init()
}

func TestPermission7(t *testing.T) {
	database.Init()
	a, err := functions.DoesUserHaveActiveSubscription(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(a)
}

// main_test.go
