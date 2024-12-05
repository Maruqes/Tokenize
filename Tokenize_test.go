package Tokenize

import (
	"testing"

	"github.com/Maruqes/Tokenize/database"
)

func TestThis(t *testing.T) {
	Init("4242", "/success.html", "/cancel.html", TypeOfSubscriptionValues.OnlyStartOnDayXNoSubscription, []ExtraPayments{ExtraPaymentsValues.Multibanco})
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
	getLastTimeOffline(1)
}

func TestPermission7(t *testing.T) {
	database.Init()
	a, err := DoesUserHaveActiveSubscription(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(a)
}

// main_test.go

func TestGetLastTimeAlgorithm(t *testing.T) {
	// Set the environment variable for the test

	// Test cases with 1 month subscription

	tests := []struct {
		name            string
		offlinePayments []database.OfflinePayment
		expectedExpiry  database.Date
	}{
		{
			name: "Single payment on 1/1 for quantity 1",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 2,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 3,
				Day:   1,
			},
		},
		{
			name: "Payment on 1/1 and another payment on 1/15 for quantities 1 and 2",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   15,
					},
					Quantity: 2,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 4,
				Day:   1,
			},
		},
		{
			name: "Payment on 1/1, current date is 1/4 (should return expiry of 1/2)",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 2,
				Day:   1,
			},
		},
		{
			name: "Two payments on 1/1 for quantity 1 each",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 3,
				Day:   1,
			},
		},
		{
			name: "Payment on 1/1 and another on 2/1 for quantity 1 each",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 2,
						Day:   1,
					},
					Quantity: 1,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 3,
				Day:   1,
			},
		},
		{
			name: "Payment on 1/1 and another on 3/1 (after expiry)",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 3,
						Day:   1,
					},
					Quantity: 1,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 4,
				Day:   1,
			},
		},
		{
			name: "Payment on 1/1 with quantity 2",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 2,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 3,
				Day:   1,
			},
		},
		{
			name: "Payment on 1/1, another payment on 1/20 with quantity 1",
			offlinePayments: []database.OfflinePayment{
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   1,
					},
					Quantity: 1,
				},
				{
					DateOfPayment: database.Date{
						Year:  2021,
						Month: 1,
						Day:   20,
					},
					Quantity: 1,
				},
			},
			expectedExpiry: database.Date{
				Year:  2021,
				Month: 3,
				Day:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastTime, err := getLastTimeAlgorithm(tt.offlinePayments)
			if err != nil {
				t.Fatalf("Error in getLastTimeAlgorithm: %v", err)
			}

			got := database.Date{
				Day:   lastTime.Day(),
				Month: int(lastTime.Month()),
				Year:  lastTime.Year(),
			}

			if got != tt.expectedExpiry {
				t.Errorf("Expected expiry date %v, but got %v", tt.expectedExpiry, got)
			}
		})
	}
}
