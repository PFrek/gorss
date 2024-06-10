package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
)

type ResponsePost struct {
	ID          uuid.UUID      `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Title       string         `json:"title"`
	Url         string         `json:"url"`
	Description sql.NullString `json:"description"`
	PublishedAt time.Time      `json:"published_at"`
	FeedID      uuid.UUID      `json:"feed_id"`
}

func postFromDBPost(post database.Post) ResponsePost {
	return ResponsePost{
		ID:          post.ID,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Title:       post.Title,
		Url:         post.Url,
		Description: post.Description,
		PublishedAt: post.PublishedAt,
		FeedID:      post.FeedID,
	}
}

func (config *ApiConfig) GetPostsHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	limitQuery, err := extractQuery(req, "limit")
	if err != nil {
		limitQuery = "10"
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		limit = 10
	}

	ctx := req.Context()

	posts, err := config.DB.GetPostsByUser(ctx, database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, 500, "Failed to get posts")
		return
	}

	response := []ResponsePost{}
	for _, post := range posts {
		response = append(response, postFromDBPost(post))
	}

	respondWithJSON(w, 200, response)

}
