package main

import (
	"log"
	"net/http"
	"os"

	"github.com/PFrek/gorss/api"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: mux,
	}

	mux.HandleFunc("/v1/healthz", api.GetHealthzHandler)
	mux.HandleFunc("/v1/err", api.GetErrorHandler)

	log.Printf("Starting server on port %s\n", port)
	server.ListenAndServe()
}
