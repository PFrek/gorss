package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ResponseFeedFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FeedID    uuid.UUID `json:"feed_id"`
	UserID    uuid.UUID `json:"user_id"`
}

func feedFollowFromDBFeedFollow(feedFollow database.FeedFollow) ResponseFeedFollow {
	return ResponseFeedFollow{
		ID:        feedFollow.ID,
		CreatedAt: feedFollow.CreatedAt,
		UpdatedAt: feedFollow.UpdatedAt,
		FeedID:    feedFollow.FeedID,
		UserID:    feedFollow.UserID,
	}
}

func (config *ApiConfig) PostFeedFollowsHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	type parameters struct {
		FeedId uuid.UUID `json:"feed_id"`
	}

	reqBody := parameters{}
	err := extractBody(req, &reqBody)
	if err != nil || reqBody.FeedId.String() == "" {
		respondWithError(w, 400, "Invalid request body")
		return
	}

	ctx := req.Context()

	_, err = config.DB.GetFeed(ctx, reqBody.FeedId)
	if err != nil {
		respondWithError(w, 404, "Feed Not Found")
		return
	}

	currentTime := time.Now().UTC()
	feedFollow, err := config.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		FeedID:    reqBody.FeedId,
		UserID:    user.ID,
	})
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				respondWithError(w, 400, "Feed already followed by user")
				return
			}
		}

		respondWithError(w, 500, fmt.Sprintf("Failed to create feed follow: %v", err))
		return
	}

	response := feedFollowFromDBFeedFollow(feedFollow)
	respondWithJSON(w, 201, response)
}

func (config *ApiConfig) DeleteFeedFollowHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	idStr := req.PathValue("feedFollowID")
	feedFollowID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, 400, "Invalid FeedFollow ID")
		return
	}

	ctx := req.Context()

	feedFollow, err := config.DB.GetFeedFollow(ctx, feedFollowID)
	if err != nil {
		respondWithError(w, 404, "FeedFollow Not Found")
		return
	}

	if feedFollow.UserID != user.ID {
		respondWithError(w, 403, "Forbidden")
		return
	}

	err = config.DB.DeleteFeedFollow(ctx, feedFollowID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed to delete Feed Follow: %v", err))
		return
	}

	w.WriteHeader(204)
}

func (config *ApiConfig) GetFeedFollowsHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	ctx := req.Context()
	feedFollows, err := config.DB.GetFeedFollows(ctx, user.ID)
	if err != nil {
		respondWithError(w, 404, "Not Found")
		return
	}

	response := []ResponseFeedFollow{}
	for _, feedFollow := range feedFollows {
		response = append(response, feedFollowFromDBFeedFollow(feedFollow))
	}

	respondWithJSON(w, 200, response)
}
