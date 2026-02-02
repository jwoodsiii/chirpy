package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jwoodsiii/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = cfg.db.RevokeToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusNoContent, "")

}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type responseBody struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	log.Printf("Bearer token: %s", token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil || rToken.Token == "" {
		log.Printf("Error pulling token from db: %v", err)
		respondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	jwt, err := auth.MakeJWT(rToken.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, responseBody{
		Token: jwt,
	})

}
