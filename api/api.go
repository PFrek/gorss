package api

import (
	"encoding/json"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Failed to marshal payload")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	body := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	data, err := json.Marshal(body)
	if err != nil {
		respondWithError(w, 500, "Failed to marshal payload")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func GetHealthzHandler(w http.ResponseWriter, req *http.Request) {
	response := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	respondWithJSON(w, 200, response)
}

func GetErrorHandler(w http.ResponseWriter, req *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}
