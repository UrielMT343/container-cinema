package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func Respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	payload, err := json.Marshal(Response{Data: data})
	if err != nil {
		log.Printf("marshal error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	_, errWrite := w.Write(payload)
	if errWrite != nil {
		log.Printf("failed to write response to client: %v", err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	Respond(w, status, Response{Error: message})
}
