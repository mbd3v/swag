package main

import "net/http"

// @title Swagger Webhooks Example API
// @version 1.0
// @description Exercises the @Webhook annotation alongside a normal @Router endpoint.
func main() {
	http.HandleFunc("/pets", ListPets)
}

// @Summary List pets
// @Success 200 {array} Pet
// @Router /pets [get]
func ListPets(w http.ResponseWriter, r *http.Request) {
}

// @Summary New pet posted
// @Description Sent when a new pet is added to the store.
// @Accept json
// @Param pet body Pet true "the new pet"
// @Success 200
// @Webhook newPetPosted [post]
func NewPetPosted() {
}

// @Summary Pet deleted
// @Description Sent when a pet is removed from the store.
// @Accept json
// @Param pet body Pet true "the deleted pet"
// @Success 200
// @Webhook petDeleted [post]
func PetDeleted() {
}

type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
