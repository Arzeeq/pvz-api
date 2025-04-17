package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPasswordWithoutError(t *testing.T) {
	password := "some_password"

	_, err := HashPassword(password)

	require.NoError(t, err)
}

func TestComparePasswords(t *testing.T) {
	password := "some_password"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	require.True(t, ComparePasswords(hash, password))
}
