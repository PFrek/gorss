package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PFrek/gorss/internal/api"
	"github.com/PFrek/gorss/internal/database"
	"github.com/PFrek/gorss/internal/scraper"
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

	// Scraper
	scraper := scraper.Scraper{
		Config:        apiConfig,
		Cache:         make(map[string]scraper.CachedFeed),
		CacheInterval: 30 * time.Minute,
	}
	scraper.Start(60*time.Second, 10)

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

	mux.HandleFunc("GET /v1/posts", apiConfig.MiddleWareAuth(apiConfig.GetPostsHandler))

	log.Printf("Starting server on port %s\n", port)
	server.ListenAndServe()

}
