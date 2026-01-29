package main

import (
	"net/http"
)

func main() {
	// http multiplexer that matches url of requests against list of registered patterns and calls handler for closest match
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	server.ListenAndServe()

}
