package server

import (
	"context"
	"testing"
	"time"

	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Run("test server whitout opts", func(t *testing.T) {
		s, err := NewServer()
		require.NoError(t, err)
		require.Empty(t, s.grpsServer)
		require.Empty(t, s.httpServer)
		require.Empty(t, s.Listener)
	})

	t.Run("test server with HTTP", func(t *testing.T) {
		httpOpt := WithHTTP(nil, nil, nil, nil, parameters.ServerParameters{
			CryptoKeyPath: "./testdata/test_private",
		})

		s, err := NewServer(httpOpt)
		require.NoError(t, err)
		require.Empty(t, s.grpsServer)
		require.NotEmpty(t, s.httpServer)
		require.Empty(t, s.Listener)
	})

	t.Run("test error server with HTTP", func(t *testing.T) {
		httpOpt := WithHTTP(nil, nil, nil, nil, parameters.ServerParameters{
			CryptoKeyPath: "./testdata/error",
		})

		_, err := NewServer(httpOpt)
		require.Error(t, err)
	})

	t.Run("test server with GRPC", func(t *testing.T) {
		grpcOpt := WithGRPC(nil, nil, nil, parameters.ServerParameters{
			FlagRunGRPCAddr: "localhost:0",
		})

		s, err := NewServer(grpcOpt)
		require.NoError(t, err)
		require.NotEmpty(t, s.grpsServer)
		require.Empty(t, s.httpServer)
		require.NotEmpty(t, s.Listener)
	})

	t.Run("test error server with GRPC", func(t *testing.T) {
		grpcOpt := WithGRPC(nil, nil, nil, parameters.ServerParameters{
			FlagRunGRPCAddr: "error",
		})

		_, err := NewServer(grpcOpt)
		require.Error(t, err)
	})

	t.Run("test server with HTTP and GRPC", func(t *testing.T) {
		httpOpt := WithHTTP(nil, nil, nil, nil, parameters.ServerParameters{
			CryptoKeyPath: "./testdata/test_private",
		})

		grpcOpt := WithGRPC(nil, nil, nil, parameters.ServerParameters{
			FlagRunGRPCAddr: "localhost:0",
		})

		s, err := NewServer(httpOpt, grpcOpt)
		require.NoError(t, err)
		require.NotEmpty(t, s.grpsServer)
		require.NotEmpty(t, s.httpServer)
		require.NotEmpty(t, s.Listener)
	})
}

func TestServer_Run(t *testing.T) {
	t.Run("test run HTTP", func(t *testing.T) {
		t.Parallel()
		httpOpt := WithHTTP(nil, nil, nil, nil, parameters.ServerParameters{
			CryptoKeyPath: "./testdata/test_private",
			FlagRunAddr:   ":0",
		})

		s, err := NewServer(httpOpt)
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-time.After(5 * time.Second)
			cancel()
		}()

		err = s.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("test run grpc", func(t *testing.T) {
		t.Parallel()
		grpcOpt := WithGRPC(nil, nil, nil, parameters.ServerParameters{
			FlagRunGRPCAddr: ":1",
		})

		s, err := NewServer(grpcOpt)
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-time.After(5 * time.Second)
			cancel()
		}()

		err = s.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("test run grpc and http", func(t *testing.T) {
		t.Parallel()
		httpOpt := WithHTTP(nil, nil, nil, nil, parameters.ServerParameters{
			CryptoKeyPath: "./testdata/test_private",
			FlagRunAddr:   ":2",
		})

		grpcOpt := WithGRPC(nil, nil, nil, parameters.ServerParameters{
			FlagRunGRPCAddr: ":3",
		})

		s, err := NewServer(httpOpt, grpcOpt)
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-time.After(5 * time.Second)
			cancel()
		}()

		err = s.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("test run HTTP error", func(t *testing.T) {
		t.Parallel()
		httpOpt := WithHTTP(nil, nil, nil, nil, parameters.ServerParameters{
			CryptoKeyPath: "./testdata/test_private",
			FlagRunAddr:   ":4",
		})

		grpcOpt := WithGRPC(nil, nil, nil, parameters.ServerParameters{
			FlagRunGRPCAddr: ":5",
		})

		s, err := NewServer(httpOpt, grpcOpt)
		require.NoError(t, err)

		go func() {
			<-time.After(1 * time.Second)
			s.Listener.Close()
		}()

		err = s.Run(context.Background())
		require.Error(t, err)
	})
}
