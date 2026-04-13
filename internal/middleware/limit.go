package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"

	"start/internal/response"

	"golang.org/x/time/rate"
)

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func EnforceSizeLimit(next http.Handler, limit int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > limit {
			slog.Warn("Request rejected by middleware: content-length too large", "size", r.ContentLength)
			w.Header().Set("Connection", "close")
			response.Error(w, http.StatusRequestEntityTooLarge, "payload exceeds size limit")
			return
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}

		next.ServeHTTP(w, r)
	})
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(2, 5)
		visitors[ip] = limiter
	}
	return limiter
}

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		ipHost, _, err := net.SplitHostPort(ip)
		if err != nil {
			slog.Error("Error spliting the host port", "error", err)
			response.Error(w, http.StatusInternalServerError, "an unexpected error ocurred")
			return
		}

		limiter := getVisitor(ipHost)

		if !limiter.Allow() {
			slog.Warn("Request rejected by middleware: too many requests", "ip", ip)
			response.Error(w, http.StatusTooManyRequests, "too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}
