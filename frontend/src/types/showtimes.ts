export interface MovieSummary {
  id: number;
  name: string;
}

export interface AuditoriumSummary {
  id: number;
  name: string;
}

export interface Showtime {
  id: number;
  startTime: string;
  movie: MovieSummary;
  auditorium: AuditoriumSummary;
}

export interface ErrorResponse {
  error: string;
}
