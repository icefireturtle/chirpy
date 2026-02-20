package main

import (
	"encoding/json"
	"net/http"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Body string `json:"body"`
	}
	type returnValue struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	p := parameter{}

	err := decoder.Decode(&p)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(p.Body) > 140 {
		errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	jsonResponse(w, http.StatusOK, returnValue{Valid: true})
}
