package server

import (
	"context"
	"testing"

	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	t.Run("test error db rep", func(t *testing.T) {
		_, err := NewRepository(context.Background(), parameters.ServerParameters{
			DataBaseDSN: "123",
		})

		require.Error(t, err)
	})

	t.Run("test in memory storage", func(t *testing.T) {
		ms, err := storage.NewMemStorage(context.Background(), parameters.ServerParameters{})
		require.NoError(t, err)

		r, err := NewRepository(context.Background(), parameters.ServerParameters{})

		require.NoError(t, err)
		require.IsType(t, ms, r)
	})

	t.Run("test in memory storage error", func(t *testing.T) {
		_, err := NewRepository(context.Background(), parameters.ServerParameters{
			FileStoragePath: "//",
		})

		require.Error(t, err)
	})
}
