package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jwoodsiii/chirpy/internal/auth"
	"github.com/jwoodsiii/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUserUpdates(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseBody struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	var params requestBody
	if err := json.Unmarshal(dat, &params); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{ID: userID, Email: params.Email, HashedPassword: hashed})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, responseBody{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseBody struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't read request")
		return
	}

	var params requestBody
	if err := json.Unmarshal(dat, &params); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't unmarshal request")
		return
	}

	expirationTime := time.Hour

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	exists, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !exists {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT")
		return
	}

	refresh, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Auth error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}

	refreshToken, err := cfg.db.CreateToken(r.Context(), database.CreateTokenParams{Token: refresh, UserID: user.ID})
	if err != nil {
		log.Printf("Database error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}

	respondWithJson(w, http.StatusOK, responseBody{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refreshToken.Token,
	})
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseBody struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't read request")
		return
	}

	var params requestBody
	if err := json.Unmarshal(dat, &params); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't unmarshal request")
		return
	}

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't hash password")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashed})
	if err != nil {
		// log.Printf("DB error: %v", err)
		respondWithError(w, http.StatusBadRequest, "couldn't create user")
		fmt.Printf("DB error: %v", err)
		return
	}

	respondWithJson(w, 201, responseBody{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
