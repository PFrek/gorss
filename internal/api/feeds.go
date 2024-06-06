package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (config *ApiConfig) PostFeedsHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	reqBody := parameters{}
	err := extractBody(req, &reqBody)
	if err != nil || reqBody.Name == "" || reqBody.Url == "" {
		respondWithError(w, 400, "Invalid request body")
		return
	}

	ctx := req.Context()
	currentTime := time.Now().UTC()
	feed, err := config.DB.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      reqBody.Name,
		Url:       reqBody.Url,
		UserID:    user.ID,
	})
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				respondWithError(w, 400, "URL already registered")
				return
			}
		}
		respondWithError(w, 500, fmt.Sprintf("Failed to create feed: %v", err))
		return
	}

	respondWithJSON(w, 201, feed)
}

func (config *ApiConfig) GetFeedsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	feeds, err := config.DB.GetFeeds(ctx)
	if err != nil {
		respondWithError(w, 500, "Failed to get feeds")
		return
	}

	respondWithJSON(w, 200, feeds)
}
