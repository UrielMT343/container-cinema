package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewConnection(ctx context.Context, connString string) (*Service, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error while parsing the database config: %v", err)
	}

	// Configuration
	config.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping the database: %v", err)
	}

	return &Service{db: pool}, nil
}

func (s *Service) CloseConnection() {
	s.db.Close()
}

func (s *Service) GetDB() *pgxpool.Pool {
	return s.db
}

func QueryRows[T any](s *Service, ctx context.Context, query string, args ...any) ([]T, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error performing the query: %v", err)
	}

	collectedRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, fmt.Errorf("error while collecting the rows: %v", err)
	}

	return collectedRows, nil
}

func (s *Service) Health(ctx context.Context) (map[string]string, error) {
	err := s.db.Ping(ctx)
	status := make(map[string]string)

	if err != nil {
		status["status"] = "down"
		return status, fmt.Errorf("failed to ping the database: %v", err)
	}

	status["status"] = "up"
	return status, nil
}
