package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Valid bool   `json:"valid, omitempty"`
		Error string `json:"error, omitempty"`
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
	// handle chirp
	if len(params.Body) > 140 {
		respondWithError(w, 400, "chirp is too long")
		return
	}

	respondWithJson(w, 200, responseBody{
		Valid: true,
	})

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
