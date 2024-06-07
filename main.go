package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/PFrek/gorss/internal/api"
	"github.com/PFrek/gorss/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbUrl := os.Getenv("CONNECTION")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	apiConfig := api.ApiConfig{
		DB: dbQueries,
	}

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: mux,
	}

	mux.HandleFunc("GET /v1/healthz", api.GetHealthzHandler)
	mux.HandleFunc("GET /v1/err", api.GetErrorHandler)

	mux.HandleFunc("POST /v1/users", apiConfig.PostUsersHandler)
	mux.HandleFunc("GET /v1/users", apiConfig.MiddleWareAuth(apiConfig.GetCurrentUserHandler))

	mux.HandleFunc("POST /v1/feeds", apiConfig.MiddleWareAuth(apiConfig.PostFeedsHandler))
	mux.HandleFunc("GET /v1/feeds", apiConfig.GetFeedsHandler)

	mux.HandleFunc("POST /v1/feed_follows", apiConfig.MiddleWareAuth(apiConfig.PostFeedFollowsHandler))
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", apiConfig.MiddleWareAuth(apiConfig.DeleteFeedFollowHandler))
	mux.HandleFunc("GET /v1/feed_follows", apiConfig.MiddleWareAuth(apiConfig.GetFeedFollowsHandler))

	log.Printf("Starting server on port %s\n", port)
	server.ListenAndServe()
}
