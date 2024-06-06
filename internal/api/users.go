package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
)

func (config *ApiConfig) PostUsersHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}

	reqBody := parameters{}

	err := extractBody(req, &reqBody)
	if err != nil || reqBody.Name == "" {
		respondWithError(w, 400, "Invalid request body")
		return
	}

	currentTime := time.Now().UTC()
	user, err := config.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      reqBody.Name,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	respondWithJSON(w, 201, user)
}

func (config *ApiConfig) GetCurrentUserHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	respondWithJSON(w, 200, user)
}
