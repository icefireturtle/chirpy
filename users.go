package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"workspace/github.com/icefireturtle/chirpy/internal/auth"
	"workspace/github.com/icefireturtle/chirpy/internal/database"
)

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type newUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	n := newUser{}

	err := decoder.Decode(&n)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	hashed, err := auth.HashPassword(n.Password)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.queries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          n.Email,
		HashedPassword: hashed,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	response := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	jsonResponse(w, http.StatusCreated, response)

}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type User struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	creds := credentials{}

	err := decoder.Decode(&creds)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := cfg.queries.GetUserByEmail(r.Context(), creds.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		errorResponse(w, http.StatusInternalServerError, "Internal Failure")
		return
	}

	match, err := auth.CheckPasswordHash(creds.Password, user.HashedPassword)
	if err != nil || !match {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	JWTexpiration := time.Hour

	token, err := auth.MakeJWT(user.ID, cfg.JWT, JWTexpiration)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	var defaultRefreshExpiration = time.Now().UTC().Add(time.Hour * 24 * 60)

	refresh_token := auth.MakeRefreshToken()

	params := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		ExpiresAt: defaultRefreshExpiration,
		UserID:    user.ID,
	}

	session, err := cfg.queries.CreateRefreshToken(r.Context(), params)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	response := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: session.Token,
	}

	jsonResponse(w, http.StatusOK, response)

}

func (cfg *apiConfig) accountHandler(w http.ResponseWriter, r *http.Request) {

	type Account struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Missing or Invalid token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	a := Account{}

	err = decoder.Decode(&a)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	hashed, err := auth.HashPassword(a.Password)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	params := database.UpdateUserParams{
		ID:             user,
		Email:          a.Email,
		HashedPassword: hashed,
	}

	updated, err := cfg.queries.UpdateUser(r.Context(), params)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response := Account{
		Email: updated.Email,
	}

	jsonResponse(w, http.StatusOK, response)

}
