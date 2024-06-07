package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ResponseFeed struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

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

	responseFeed := ResponseFeed{
		ID:        feed.ID,
		CreatedAt: feed.CreatedAt,
		UpdatedAt: feed.UpdatedAt,
		Name:      feed.Name,
		Url:       feed.Url,
		UserID:    feed.UserID,
	}

	feedFollow, err := config.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		FeedID:    feed.ID,
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed to create feed follow: %v", err))
		return
	}

	responseFeedFollow := ResponseFeedFollow{
		ID:        feedFollow.ID,
		CreatedAt: feedFollow.CreatedAt,
		UpdatedAt: feedFollow.UpdatedAt,
		FeedID:    feedFollow.FeedID,
		UserID:    feedFollow.UserID,
	}

	response := struct {
		Feed       ResponseFeed       `json:"feed"`
		FeedFollow ResponseFeedFollow `json:"feed_follow"`
	}{
		Feed:       responseFeed,
		FeedFollow: responseFeedFollow,
	}
	respondWithJSON(w, 201, response)
}

func (config *ApiConfig) GetFeedsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	feeds, err := config.DB.GetFeeds(ctx)
	if err != nil {
		respondWithError(w, 500, "Failed to get feeds")
		return
	}

	response := []ResponseFeed{}
	for _, feed := range feeds {
		response = append(response, ResponseFeed{
			ID:        feed.ID,
			CreatedAt: feed.CreatedAt,
			UpdatedAt: feed.UpdatedAt,
			Name:      feed.Name,
			Url:       feed.Url,
			UserID:    feed.UserID,
		})
	}
	respondWithJSON(w, 200, response)
}
