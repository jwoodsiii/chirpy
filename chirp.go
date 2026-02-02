package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jwoodsiii/chirpy/internal/auth"
	"github.com/jwoodsiii/chirpy/internal/database"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 403, err.Error())
		return
	}

	id := r.PathValue("chirpID")

	chirp, err := cfg.db.GetChirp(r.Context(), uuid.MustParse(id))
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "not authorized to delete this chirp")
	}
	_, err = cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{ID: uuid.MustParse(id), UserID: userID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusNoContent, "")

}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := r.PathValue("chirpID")
	log.Printf("ID before uuid parse: %s", id)
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "missing chirp ID")
		return
	}

	type responseBody struct {
		Chirp
	}

	chirp, err := cfg.db.GetChirp(r.Context(), uuid.MustParse(id))
	if err != nil {
		log.Printf("Database error: %v attempting to pull id: %s", err, uuid.MustParse(id))
		respondWithError(w, http.StatusNotFound, "Couldn't retrieve chirp")
		return
	}

	respondWithJson(w, http.StatusOK, responseBody{
		Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	author := r.URL.Query().Get("author_id")
	log.Printf("author id: %s", author)

	chirps := []Chirp{}
	if author != "" {
		dbChirps, err := cfg.db.GetChirpsByAuthor(r.Context(), uuid.MustParse(author))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
			return
		} else {
			for _, dbChirp := range dbChirps {
				chirps = append(chirps, Chirp{
					Id:        dbChirp.ID,
					CreatedAt: dbChirp.CreatedAt,
					UpdatedAt: dbChirp.UpdatedAt,
					UserId:    dbChirp.UserID,
					Body:      dbChirp.Body,
				})
			}
			respondWithJson(w, http.StatusOK, chirps)
			return
		}
	}

	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			Id:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserId:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Chirp
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	userId, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
		return
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

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: handleProfanity(params.Body), UserID: userId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp")
		return
	}

	// handle profane
	respondWithJson(w, http.StatusCreated, responseBody{
		Chirp{
			Id:        chirp.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	})
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
