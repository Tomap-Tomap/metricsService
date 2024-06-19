package client

import (
	"testing"

	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("positive grpc", func(t *testing.T) {
		p := parameters.AgentParameters{
			HashKey:    "",
			RateLimit:  1,
			ListenAddr: ":0",
			UseGRPC:    true,
		}
		gc, err := NewGRPC(p)
		require.NoError(t, err)

		c, err := NewClient(p)
		require.NoError(t, err)
		require.IsType(t, gc, c)
	})

	t.Run("positive http", func(t *testing.T) {
		p := parameters.AgentParameters{
			CryptoKeyPath: "./testdata/test_public",
			HashKey:       "",
			RateLimit:     1,
			ListenAddr:    ":0",
			UseGRPC:       false,
		}
		gc, err := NewHTTP(p)
		require.NoError(t, err)

		c, err := NewClient(p)
		require.NoError(t, err)
		require.IsType(t, gc, c)
	})

	t.Run("negative http", func(t *testing.T) {
		p := parameters.AgentParameters{
			CryptoKeyPath: "",
			HashKey:       "",
			RateLimit:     1,
			ListenAddr:    ":0",
			UseGRPC:       false,
		}

		_, err := NewClient(p)
		require.Error(t, err)
	})
}
