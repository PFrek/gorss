package api

import (
	"errors"
	"net/http"
	"strings"
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
