package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"workspace/github.com/icefireturtle/chirpy/internal/auth"
	"workspace/github.com/icefireturtle/chirpy/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type incomingChirp struct {
		Body string `json:"body"`
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	const maxBodyLength = 140

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	i := incomingChirp{}

	err = decoder.Decode(&i)
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
		UserID: userID,
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

	s := r.URL.Query().Get("author_id")
	t := r.URL.Query().Get("sort")

	var stored []database.Chirp
	var err error

	if s != "" {
		q, err := uuid.Parse(s)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed parse")
			return
		}
		stored, err = cfg.queries.ViewChirpsByAuthor(r.Context(), q)
		if err != nil {
			log.Printf("Error viewing chirps: %v", err)
			errorResponse(w, http.StatusInternalServerError, "Failed to retreive chirps")
			return
		}
	} else if t != "" {
		if t == "desc" {
			stored, err = cfg.queries.ViewChirps(r.Context())
			if err != nil {
				log.Printf("Error viewing chirps: %v", err)
				errorResponse(w, http.StatusInternalServerError, "Failed to retreive chirps")
				return
			}
			sort.Slice(stored, func(i, j int) bool { return stored[i].CreatedAt.After(stored[j].CreatedAt) })
		} else {
			stored, err = cfg.queries.ViewChirps(r.Context())
			if err != nil {
				log.Printf("Error viewing chirps: %v", err)
				errorResponse(w, http.StatusInternalServerError, "Failed to retreive chirps")
				return
			}
		}
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

func (cfg *apiConfig) viewChirpHandler(w http.ResponseWriter, r *http.Request) {

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	single, err := cfg.queries.ViewChirp(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse(w, http.StatusNotFound, "Chirp not found")
			return
		}
		log.Printf("Error viewing chirp: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to retreive chirp")
		return
	}

	response := Chirp{
		ID:        single.ID,
		CreatedAt: single.CreatedAt,
		UpdatedAt: single.UpdatedAt,
		Body:      single.Body,
		UserID:    single.UserID,
	}

	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	user, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Missing or Invalid token")
		return
	}

	chirp, err := cfg.queries.ViewChirp(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse(w, http.StatusNotFound, "Chirp not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, "Failed to retreive chirp")
		return
	}

	if chirp.UserID != user {
		errorResponse(w, http.StatusForbidden, "Forbidden")
		return
	}

	params := database.DeleteChirpParams{
		ID:     chirp.ID,
		UserID: user,
	}

	_, err = cfg.queries.DeleteChirp(r.Context(), params)
	if err != nil {
		errorResponse(w, http.StatusForbidden, "Forbidden")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
