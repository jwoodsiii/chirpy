package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {

	apiConfig := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	const filePathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app", apiConfig.middlewareMetricsInc(http.FileServer(http.Dir(filePathRoot)))))
	// mux.Handle("/assets/logo.png", http.FileServer(http.Dir(".")))
	// custom handler for readiness endpoint
	mux.HandleFunc("GET /healthz", handlerReadiness)
	mux.HandleFunc("GET /metrics", apiConfig.handlerRequestCounter)
	mux.HandleFunc("POST /reset", apiConfig.handlerReset)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	server.ListenAndServe()

}

func (cfg *apiConfig) handlerRequestCounter(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + strconv.Itoa(int(cfg.fileserverHits.Load()))))
}
