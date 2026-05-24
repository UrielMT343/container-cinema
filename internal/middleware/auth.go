package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"start/internal/auth"
	redisclient "start/internal/redis"
	"start/internal/response"
)

func AdminAuth(tokenSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCookie, err := r.Cookie("cinema_auth_token")
			if err != nil {
				slog.Warn("Request rejected by middleware: no cookie found", "error", err)
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			tokenString := authCookie.Value

			tokenClaims, err := auth.ValidateToken(tokenString, tokenSecret)
			if err != nil {
				slog.Warn("Request rejected by middleware: token not validated", "error", err)
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if tokenClaims.Role != "admin" {
				slog.Warn("Request rejected by middleware: invalid role", "error", err)
				response.Error(w, http.StatusForbidden, "Forbidden")
				return
			}

			if !tokenClaims.Authorized {
				slog.Warn("Request rejected by middleware: not authorized", "error", err)
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CartAuth(redis *redisclient.Redis) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("cart_id")
			if err != nil {
				slog.Warn("Request rejected by middleware: no cookie found", "error", err)
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			cartID := cookie.Value

			exists, err := redis.Exists(r.Context(), cartID)
			if err != nil {
				slog.Warn("Request rejected by middleware: could not found cartID on registry", "error", err)
				response.Error(w, http.StatusInternalServerError, "Unauthorized")
				return
			}

			if exists == 0 {
				slog.Warn("Request rejected by middleware: could not found cartID on registry", "exists", exists)
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), auth.CartContextKey, cartID)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
