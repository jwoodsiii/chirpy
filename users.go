package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
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

	var input requestBody
	if err := json.Unmarshal(dat, &input); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't unmarshal request")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), input.Email)
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
