package main

import "net/http"

// @title Schema Keywords Example API
// @version 1.0
// @description Exercises the last two JSON Schema 2020-12 keywords: patternProperties and propertyNames.
func main() {
	http.HandleFunc("/configs", GetConfig)
}

// @Summary Get a config
// @Success 200 {object} Config
// @Router /configs [get]
func GetConfig(w http.ResponseWriter, r *http.Request) {
}

type Config struct {
	// Extra holds dynamically-named settings; a key's prefix determines its value type.
	Extra  map[string]any `json:"extra" patternProperties:"^is_.*=bool,^count_.*=int"`
	Labels Labels         `json:"labels"`
}

// Labels holds arbitrary key/value tags whose keys must be lowercase snake_case.
// @PropertyNames ^[a-z][a-z0-9_]*$
type Labels map[string]string
