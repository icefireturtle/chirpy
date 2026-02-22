package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"workspace/github.com/icefireturtle/chirpy/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type incomingChirp struct {
		Body    string    `json:"body"`
		User_id uuid.UUID `json:"user_id"`
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	const maxBodyLength = 140

	decoder := json.NewDecoder(r.Body)
	i := incomingChirp{}

	err := decoder.Decode(&i)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(i.Body) > maxBodyLength {
		errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	for _, profane := range strings.Fields(i.Body) {

		lower := strings.ToLower(profane)

		switch lower {
		case "kerfuffle", "sharbert", "fornax":
			i.Body = strings.ReplaceAll(i.Body, profane, "****")
		}

	}

	params := database.CreateChirpsParams{
		Body:   i.Body,
		UserID: i.User_id,
	}

	chirp, err := cfg.queries.CreateChirps(r.Context(), params)
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	response := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	jsonResponse(w, http.StatusCreated, response)
}

func (cfg *apiConfig) viewChirpsHandler(w http.ResponseWriter, r *http.Request) {

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	stored, err := cfg.queries.ViewChirps(r.Context())
	if err != nil {
		log.Printf("Error viewing chirps: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to view chirps")
		return
	}

	chirps := []Chirp{}

	for _, c := range stored {
		chirps = append(chirps, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}

	jsonResponse(w, http.StatusOK, chirps)
}
