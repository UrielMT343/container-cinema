package main

import (
	"net/http"
	"start/internal/movie"
	"start/internal/seat"
	"start/internal/showtime"
	"start/internal/ticket"
)

type Config struct {
	movieHanlder    *movie.Handler
	seatHandler     *seat.Hander
	showtimeHanlder *showtime.Handler
	ticketHandler   *ticket.Handler
}

func routes(c *Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /movies", c.movieHanlder.GetMovies)
	mux.HandleFunc("POST /movies", c.movieHanlder.InsertMovie)

	mux.HandleFunc("GET /seats", c.seatHandler.GetSeats)
	mux.HandleFunc("POST /seats", c.seatHandler.InsertSeat)
	mux.HandleFunc("GET /seats/auditorium/{id}", c.seatHandler.GetSeatsByAuditorium)
	mux.HandleFunc("GET /seats/showtime/{id}", c.seatHandler.GetSeatsByShowtime)

	mux.HandleFunc("GET /showtimes", c.showtimeHanlder.GetShowtimes)
	mux.HandleFunc("GET /showtimes/{id}", c.showtimeHanlder.GetShowtimesById)

	mux.HandleFunc("POST /ticket", c.ticketHandler.HoldTicket)
	mux.HandleFunc("PATCH /ticket/{id}/pay", c.ticketHandler.ConfirmTicket)
	mux.HandleFunc("DELETE /ticket/{id}", c.ticketHandler.CancelTicket)
	return mux
}
