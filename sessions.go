package main

import (
	"net/http"
	"time"
	"workspace/github.com/icefireturtle/chirpy/internal/auth"
)

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {

	type Refresh struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.queries.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	refreshed, err := auth.MakeJWT(user.ID, cfg.JWT, time.Hour)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response := Refresh{
		Token: refreshed,
	}

	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	err = cfg.queries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
