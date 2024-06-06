package api

import (
	"context"
	"net/http"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
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

	currentTime := time.Now().UTC()
	feed, err := config.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      reqBody.Name,
		Url:       reqBody.Url,
		UserID:    user.ID,
	})

	respondWithJSON(w, 201, feed)
}
