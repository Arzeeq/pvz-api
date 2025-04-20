package pg

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPgxPool struct {
	mock.Mock
	pgxpool.Pool
}

func TestNewUserStorage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pool := &pgxpool.Pool{}
		storage, err := NewUserStorage(pool)
		require.NoError(t, err)
		require.NotNil(t, storage)
	})

	t.Run("nil pool", func(t *testing.T) {
		storage, err := NewUserStorage(nil)
		require.Error(t, err)
		require.Nil(t, storage)
	})
}
