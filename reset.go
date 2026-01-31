package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if cfg.platform != "dev" {
		respondWithError(w, 403, "can't delete all users in prod, are you crazy?")
		return
	}

	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		respondWithError(w, 500, "failed to delete users")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{"message": "reset successful"})
}
