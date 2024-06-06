package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/PFrek/gorss/internal/database"
)

type Authorization struct {
	Label string
	Key   string
}

func (auth Authorization) isValidApiKey() bool {
	return auth.Label == "ApiKey" && len(auth.Key) == 64
}

func getAuthorization(req *http.Request) (Authorization, error) {
	authStr := req.Header.Get("Authorization")

	parts := strings.Split(authStr, " ")
	if len(parts) != 2 {
		return Authorization{}, errors.New("Invalid number of elements in Authorization header")
	}

	return Authorization{
		Label: parts[0],
		Key:   parts[1],
	}, nil
}

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (config *ApiConfig) MiddleWareAuth(handler authedHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		auth, err := getAuthorization(req)
		if err != nil || !auth.isValidApiKey() {
			respondWithError(w, 401, "Unauthorized")
			return
		}

		ctx := req.Context()

		user, err := config.DB.GetUser(ctx, auth.Key)
		if err != nil {
			respondWithError(w, 404, "Not found")
			return
		}

		handler(w, req, user)
	})
}
