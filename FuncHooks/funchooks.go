package funchooks

import "net/http"

//there functions will be called before the father function	is called

var CreateUser_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var CreatePortalSession_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var Checkout_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var LogoutUser_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var LoginUser_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var Multibanco_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var PayOffline_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil
var IsOffline_UserFunc func(w http.ResponseWriter, r *http.Request) bool = nil

func SetCreateUser_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	CreateUser_UserFunc = f
}

func SetLoginUser_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	LoginUser_UserFunc = f
}

func SetLogoutUser_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	LogoutUser_UserFunc = f
}

// For stripe portal
func SetCreatePortalSession_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	CreatePortalSession_UserFunc = f
}

// Works for any subscription type
func SetCheckout_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	Checkout_UserFunc = f
}

// For multibanco
func SetMultibanco_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	Multibanco_UserFunc = f
}

// For offline payment
func SetPayOffline_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	PayOffline_UserFunc = f
}

func SetIsOffline_UserFunc(f func(w http.ResponseWriter, r *http.Request) bool) {
	IsOffline_UserFunc = f
}