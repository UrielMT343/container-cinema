package health

import (
	"net/http"

	"start/internal/response"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Respond(w, http.StatusOK, "PONG")
}
