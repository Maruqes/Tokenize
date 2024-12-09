package types

type TypeOfSubscription string

// these subscriptions are only available with card
// Agrupa os valores do enum num struct
var TypeOfSubscriptionValues = struct {
	Normal                        TypeOfSubscription
	OnlyStartOnDayX               TypeOfSubscription
	OnlyStartOnDayXNoSubscription TypeOfSubscription
	MourosSubscription            TypeOfSubscription
}{
	Normal:                        "Normal",
	OnlyStartOnDayX:               "OnlyStartOnDayX",
	OnlyStartOnDayXNoSubscription: "OnlyStartOnDayXNoSubscription",
	MourosSubscription:            "MourosSubscription",
}

// a subscription you need to pay manually for now with mbway/multibanco both portuguese payment methods
type ExtraPayments string

var ExtraPaymentsValues = struct {
	MBWay      ExtraPayments
	Multibanco ExtraPayments
}{
	MBWay:      "mbway",
	Multibanco: "multibanco",
}

var GLOBAL_TYPE_OF_SUBSCRIPTION = TypeOfSubscriptionValues.Normal
var GLOBAL_EXTRA_PAYMENTS = []ExtraPayments{}
