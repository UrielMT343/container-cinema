package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Data any `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func Respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(Response{Data: data})
	if err != nil {
		log.Printf("failed to write success response: %v", err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(ErrorResponse{Error: message})
	if err != nil {
		log.Printf("failed to write error response: %v", err)
	}
}
