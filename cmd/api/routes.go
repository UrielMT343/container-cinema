package main

import (
	"net/http"
	"strings"

	"start/internal/middleware"
	"start/internal/movie"
	"start/internal/seat"
	"start/internal/showtime"
	"start/internal/ticket"
	"start/internal/user"
)

type Config struct {
	movieHanlder    *movie.Handler
	seatHandler     *seat.Hander
	showtimeHanlder *showtime.Handler
	ticketHandler   *ticket.Handler
	userHandler     *user.Handler
}

func routes(c *Config, basePrefix string, secret string) http.Handler {
	mainMux := http.NewServeMux()

	apiMux := http.NewServeMux()

	userMux := http.NewServeMux()
	adminMux := http.NewServeMux()
	publicMux := http.NewServeMux()

	publicMux.HandleFunc("GET /showtimes", c.showtimeHanlder.GetShowtimes)
	publicMux.HandleFunc("GET /showtimes/{id}", c.showtimeHanlder.GetShowtimesByID)
	publicMux.HandleFunc("GET /seats/showtime/{id}", c.seatHandler.GetSeatsByShowtime)
	publicMux.HandleFunc("POST /login", c.userHandler.LoginUser)
	publicMux.HandleFunc("POST /checkout/begin", c.ticketHandler.BeginCheckout)

	adminMux.HandleFunc("GET /movies", c.movieHanlder.GetMovies)
	adminMux.HandleFunc("POST /movies", c.movieHanlder.InsertMovie)
	adminMux.HandleFunc("POST /seats", c.seatHandler.InsertSeat)
	adminMux.HandleFunc("GET /seats", c.seatHandler.GetSeats)
	adminMux.HandleFunc("GET /seats/auditorium/{id}", c.seatHandler.GetSeatsByAuditorium)
	adminMux.HandleFunc("POST /users", c.userHandler.InsertUser)

	userMux.HandleFunc("POST /ticket/hold", c.ticketHandler.HoldTicket)
	userMux.HandleFunc("PATCH /ticket/pay", c.ticketHandler.ConfirmTicket)

	prefix := basePrefix
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	publicStack := middleware.CreateChain(publicMux, middleware.RateLimiter)
	adminStack := middleware.CreateChain(adminMux, middleware.RateLimiter, middleware.AdminAuth(secret))
	userStack := middleware.CreateChain(userMux, middleware.RateLimiter, middleware.CartAuth())

	apiMux.Handle("/public/", http.StripPrefix("/public", publicStack))
	apiMux.Handle("/admin/", http.StripPrefix("/admin", adminStack))
	apiMux.Handle("/user/", http.StripPrefix("/user", userStack))

	mainMux.Handle(prefix, http.StripPrefix(basePrefix, apiMux))

	finalMux := middleware.Logger(mainMux)

	return finalMux
}
