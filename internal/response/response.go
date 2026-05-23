package response

import (
	"encoding/json"
	"log/slog"
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

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("Failed to write success response", "error", err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(ErrorResponse{Error: message})
	if err != nil {
		slog.Error("Failed to write error response", "error", err)
	}
}
