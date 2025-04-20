package pg

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestNewPVZStorage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pool := &pgxpool.Pool{}
		storage, err := NewPVZStorage(pool)
		require.NoError(t, err)
		require.NotNil(t, storage)
	})

	t.Run("nil pool", func(t *testing.T) {
		storage, err := NewPVZStorage(nil)
		require.Error(t, err)
		require.Nil(t, storage)
	})
}
