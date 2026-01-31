package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jwoodsiii/chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	type responseBody struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error reading response body")
		return
	}

	var params requestBody
	if err := json.Unmarshal(dat, &params); err != nil {
		respondWithError(w, http.StatusBadRequest, "error unmarshalling request body")
		return
	}

	// handle chirp len
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "chirp is too long")
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: handleProfanity(params.Body), UserID: params.UserId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp")
		return
	}

	// handle profane
	respondWithJson(w, 201, responseBody{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		CleanedBody string `json:"cleaned_body, omitempty"`
		Error       string `json:"error, omitempty"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		return
	}

	var params requestBody
	if err := json.Unmarshal(dat, &params); err != nil {
		respondWithError(w, 500, "couldn't unmarshal parameters")
		return
	}

}

func handleProfanity(chirp string) string {
	// kerfuffle
	// sharbert
	// fornax
	c := strings.Fields(chirp)
	for i, v := range c {
		switch strings.ToLower(v) {
		case "kerfuffle", "sharbert", "fornax":
			c[i] = "****"
		}
	}
	return strings.Join(c, " ")
}

func respondWithJson(w http.ResponseWriter, code int, payload any) error {
	resp, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(resp)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJson(w, code, map[string]string{"error": msg})
}
