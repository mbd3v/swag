package main

import "net/http"

// @title Swagger Component Reuse Example API
// @version 1.0
// @description Exercises automatic deduplication of parameters, request bodies, responses, and headers into shared components.
func main() {
	http.HandleFunc("/pets/", GetPet)
	http.HandleFunc("/widgets/", GetWidget)
	http.HandleFunc("/gadgets/", GetGadget)
	http.HandleFunc("/pets/create", CreatePet)
	http.HandleFunc("/widgets/create", CreateWidget)
}

// @Summary Get a pet
// @Param id path int true "resource ID"
// @Param X-Request-ID header string true "request tracing ID"
// @Success 200 {object} Item
// @Header 200 {string} X-RateLimit-Remaining "requests left in this window"
// @Failure 404 {object} Error "not found"
// @Router /pets/{id} [get]
func GetPet(w http.ResponseWriter, r *http.Request) {
}

// @Summary Get a widget
// @Param id path int true "resource ID"
// @Param X-Request-ID header string true "request tracing ID"
// @Success 200 {object} Item
// @Header 200 {string} X-RateLimit-Remaining "requests left in this window"
// @Failure 404 {object} Error "not found"
// @Router /widgets/{id} [get]
func GetWidget(w http.ResponseWriter, r *http.Request) {
}

// @Summary Get a gadget
// @Param id path int true "resource ID"
// @Success 200 {object} Item
// @Router /gadgets/{id} [get]
func GetGadget(w http.ResponseWriter, r *http.Request) {
}

// @Summary Create a pet
// @Accept json
// @Param item body Item true "item to create"
// @Success 200 {object} Item
// @Router /pets/create [post]
func CreatePet(w http.ResponseWriter, r *http.Request) {
}

// @Summary Create a widget
// @Accept json
// @Param item body Item true "item to create"
// @Success 200 {object} Item
// @Router /widgets/create [post]
func CreateWidget(w http.ResponseWriter, r *http.Request) {
}

type Item struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Error struct {
	Message string `json:"message"`
}
