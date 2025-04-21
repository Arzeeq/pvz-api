package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// return connection of pool, defer func, and error if is
func InitDB(connStr string) (*pgxpool.Pool, func(), error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, nil, err
	}

	return pool, func() { pool.Close() }, nil
}
