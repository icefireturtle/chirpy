package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Body string `json:"body"`
	}
	type returnValue struct {
		CleanedBody string `json:"cleaned_body"`
	}

	const maxBodyLength = 140

	decoder := json.NewDecoder(r.Body)
	p := parameter{}

	err := decoder.Decode(&p)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(p.Body) > maxBodyLength {
		errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	for _, profane := range strings.Fields(p.Body) {

		lower := strings.ToLower(profane)

		switch lower {
		case "kerfuffle", "sharbert", "fornax":
			p.Body = strings.ReplaceAll(p.Body, profane, "****")
		}

	}

	jsonResponse(w, http.StatusOK, returnValue{CleanedBody: p.Body})
}
