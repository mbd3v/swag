package main

import "net/http"

// @title Swagger Callbacks and Links Example API
// @version 1.0
// @description Exercises the @Callback and @Link annotations.
func main() {
	http.HandleFunc("/subscriptions", CreateSubscription)
	http.HandleFunc("/users/address", GetUserAddress)
}

// @Summary Create a subscription
// @ID createSubscription
// @Accept json
// @Param subscription body Subscription true "subscription request"
// @Success 200 {object} User
// @Link 200 address getUserAddress "the user's address"
// @Link.parameter address userId $response.body#/id
// @Router /subscriptions [post]
func CreateSubscription(w http.ResponseWriter, r *http.Request) {
}

// @Summary Get a user's address
// @ID getUserAddress
// @Success 200 {object} Address
// @Router /users/address [get]
func GetUserAddress(w http.ResponseWriter, r *http.Request) {
}

// @Summary New subscription data
// @Description Sent to the subscription's callback URL when new data arrives.
// @Accept json
// @Param payload body Payload true "callback payload"
// @Success 200
// @Callback createSubscription onData {$request.body#/callbackUrl} [post]
func OnDataCallback() {
}

type Subscription struct {
	CallbackURL string `json:"callbackUrl"`
}

type User struct {
	ID int64 `json:"id"`
}

type Address struct {
	Street string `json:"street"`
}

type Payload struct {
	Data string `json:"data"`
}
