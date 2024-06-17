package client

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MetricsServerMockedObject struct {
	proto.UnimplementedMetricsServer
	mock.Mock
}

func (ms *MetricsServerMockedObject) Update(context.Context, *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	args := ms.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*proto.UpdateResponse), args.Error(1)
}

func (ms *MetricsServerMockedObject) Updates(context.Context, *proto.UpdatesRequest) (*empty.Empty, error) {
	args := ms.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*empty.Empty), args.Error(1)
}

func TestGRPCClient_SendGauge(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		msmo := new(MetricsServerMockedObject)

		msmo.On("Update").Return(nil, nil)

		proto.RegisterMetricsServer(s, msmo)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		c := GRPC{
			conn:   conn,
			client: proto.NewMetricsClient(conn),
		}
		err = c.SendGauge(context.Background(), "test", 1)
		require.NoError(t, err)
	})

	t.Run("negative test", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		msmo := new(MetricsServerMockedObject)

		msmo.On("Update").Return(nil, fmt.Errorf("test error"))

		proto.RegisterMetricsServer(s, msmo)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		c := GRPC{
			conn:   conn,
			client: proto.NewMetricsClient(conn),
		}
		err = c.SendGauge(context.Background(), "test", 1)
		require.Error(t, err)
	})
}

func TestGRPCClient_SendCounter(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		msmo := new(MetricsServerMockedObject)

		msmo.On("Update").Return(nil, nil)

		proto.RegisterMetricsServer(s, msmo)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		c := GRPC{
			conn:   conn,
			client: proto.NewMetricsClient(conn),
		}
		err = c.SendCounter(context.Background(), "test", 1)
		require.NoError(t, err)
	})

	t.Run("negative test", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		msmo := new(MetricsServerMockedObject)

		msmo.On("Update").Return(nil, fmt.Errorf("test error"))

		proto.RegisterMetricsServer(s, msmo)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		c := GRPC{
			conn:   conn,
			client: proto.NewMetricsClient(conn),
		}
		err = c.SendCounter(context.Background(), "test", 1)
		require.Error(t, err)
	})
}

func TestGRPCClient_SendBatch(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		msmo := new(MetricsServerMockedObject)

		msmo.On("Updates").Return(nil, nil)

		proto.RegisterMetricsServer(s, msmo)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		c := GRPC{
			conn:   conn,
			client: proto.NewMetricsClient(conn),
		}
		err = c.SendBatch(context.Background(), map[string]float64{
			"test":  1.1,
			"test2": 2.2,
		})
		require.NoError(t, err)
	})

	t.Run("negative test", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		msmo := new(MetricsServerMockedObject)

		msmo.On("Updates").Return(nil, fmt.Errorf("test error"))

		proto.RegisterMetricsServer(s, msmo)

		go func() {
			if err := s.Serve(lis); err != nil {
				require.FailNow(t, err.Error())
			}
		}()

		defer s.Stop()

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		c := GRPC{
			conn:   conn,
			client: proto.NewMetricsClient(conn),
		}
		err = c.SendBatch(context.Background(), map[string]float64{
			"test":  1.1,
			"test2": 2.2,
		})
		require.Error(t, err)
	})
}

func TestNewGRPC(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		c, err := NewGRPC(parameters.AgentParameters{
			HashKey:    "",
			RateLimit:  1,
			ListenAddr: ":0",
		})

		require.NoError(t, err)
		defer c.Close()
		require.NotEmpty(t, c)
	})
}
