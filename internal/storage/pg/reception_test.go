package pg

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestNewReceptionStorage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pool := &pgxpool.Pool{}
		storage, err := NewReceptionStorage(pool)
		require.NoError(t, err)
		require.NotNil(t, storage)
	})

	t.Run("nil pool", func(t *testing.T) {
		storage, err := NewReceptionStorage(nil)
		require.Error(t, err)
		require.Nil(t, storage)
	})
}
