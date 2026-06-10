package middleware

import (
	"net/http"
	"strings"
	"time"
)

func RestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/sse/") {
				next.ServeHTTP(w, r)
				return
			}

			timeoutMsg := `{"error": "Request timed out, server is busy."}`
			http.TimeoutHandler(next, timeout, timeoutMsg).ServeHTTP(w, r)
		})
	}
}
