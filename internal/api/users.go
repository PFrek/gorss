package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PFrek/gorss/internal/database"
	"github.com/google/uuid"
)

type ResponseUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

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

	ctx := req.Context()
	currentTime := time.Now().UTC()
	user, err := config.DB.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      reqBody.Name,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	response := ResponseUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
	respondWithJSON(w, 201, response)
}

func (config *ApiConfig) GetCurrentUserHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	response := ResponseUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
	respondWithJSON(w, 200, response)
}
