package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
