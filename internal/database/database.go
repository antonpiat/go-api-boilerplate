package database

import (
	"context"
	"fmt"

	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabase(context context.Context, config *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// TODO: build connection string
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)

	// TODO: create connection pool
	pool, err := pgxpool.New(context, connectionString)
	if err != nil {
		return nil, err
	}

	// TODO: verify database connection
	if err := pool.Ping(context); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
