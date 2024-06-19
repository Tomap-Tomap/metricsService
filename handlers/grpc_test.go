package handlers

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestMetricsServer_Update(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()

	smo := new(StorageMockedObject)

	proto.RegisterMetricsServer(s, NewMetricsServer(smo))

	go func() {
		if err := s.Serve(lis); err != nil {
			require.FailNow(t, err.Error())
		}
	}()

	defer s.Stop()

	t.Run("positive test counter", func(t *testing.T) {
		delta := int64(1)
		smo.On("UpdateByMetrics", models.Metrics{
			Delta: &delta,
			ID:    "test",
			MType: "counter",
		}).Return(&models.Metrics{
			Delta: &delta,
			ID:    "test",
			MType: "counter",
		}, nil)

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		wantResp := &proto.UpdateResponse{
			Metric: &proto.Metric{
				Data: &proto.Metric_Delta{Delta: 1},
				Id:   "test",
				Type: proto.Types_COUNTER,
			},
		}

		resp, err := client.Update(context.Background(), &proto.UpdateRequest{
			Metric: &proto.Metric{
				Data: &proto.Metric_Delta{Delta: 1},
				Id:   "test",
				Type: proto.Types_COUNTER,
			},
		})

		require.NoError(t, err)
		require.Equal(t, wantResp.Metric, resp.Metric)
	})

	t.Run("positive test gauge", func(t *testing.T) {
		value := float64(1)
		smo.On("UpdateByMetrics", models.Metrics{
			Value: &value,
			ID:    "test",
			MType: "gauge",
		}).Return(&models.Metrics{
			Value: &value,
			ID:    "test",
			MType: "gauge",
		}, nil)

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		wantResp := &proto.UpdateResponse{
			Metric: &proto.Metric{
				Data: &proto.Metric_Value{Value: 1},
				Id:   "test",
				Type: proto.Types_GAUGE,
			},
		}

		resp, err := client.Update(context.Background(), &proto.UpdateRequest{
			Metric: &proto.Metric{
				Data: &proto.Metric_Value{Value: 1},
				Id:   "test",
				Type: proto.Types_GAUGE,
			},
		})

		require.NoError(t, err)
		require.Equal(t, wantResp.Metric, resp.Metric)
	})

	t.Run("test update error", func(t *testing.T) {
		value := float64(1)
		smo.On("UpdateByMetrics", models.Metrics{
			Value: &value,
			ID:    "error",
			MType: "gauge",
		}).Return(nil, fmt.Errorf("test error"))

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		_, err = client.Update(context.Background(), &proto.UpdateRequest{
			Metric: &proto.Metric{
				Data: &proto.Metric_Value{Value: 1},
				Id:   "error",
				Type: proto.Types_GAUGE,
			},
		})

		require.Error(t, err)
	})

	smo.AssertExpectations(t)

	t.Run("test wrong arguments", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		_, err = client.Update(context.Background(), &proto.UpdateRequest{
			Metric: &proto.Metric{
				Data: &proto.Metric_Value{Value: 1},
				Id:   "error",
				Type: proto.Types_COUNTER,
			},
		})

		require.Error(t, err)
	})
}

func TestMetricsServer_Updates(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()

	smo := new(StorageMockedObject)

	proto.RegisterMetricsServer(s, NewMetricsServer(smo))

	go func() {
		if err := s.Serve(lis); err != nil {
			require.FailNow(t, err.Error())
		}
	}()

	defer s.Stop()

	t.Run("positive test", func(t *testing.T) {
		delta := int64(1)
		smo.On("Updates", []models.Metrics{
			{
				Delta: &delta,
				ID:    "test",
				MType: "counter",
			},
		}).Return(nil)

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		_, err = client.Updates(context.Background(), &proto.UpdatesRequest{
			Metrics: []*proto.Metric{
				{
					Data: &proto.Metric_Delta{Delta: 1},
					Id:   "test",
					Type: proto.Types_COUNTER,
				},
			},
		})

		require.NoError(t, err)
	})

	t.Run("updates error", func(t *testing.T) {
		delta := int64(1)
		smo.On("Updates", []models.Metrics{
			{
				Delta: &delta,
				ID:    "error",
				MType: "counter",
			},
		}).Return(fmt.Errorf("test error"))

		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		_, err = client.Updates(context.Background(), &proto.UpdatesRequest{
			Metrics: []*proto.Metric{
				{
					Data: &proto.Metric_Delta{Delta: 1},
					Id:   "error",
					Type: proto.Types_COUNTER,
				},
			},
		})

		require.Error(t, err)
	})

	smo.AssertExpectations(t)

	t.Run("test wrong arguments", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		require.NoError(t, err)
		defer conn.Close()

		client := proto.NewMetricsClient(conn)

		_, err = client.Updates(context.Background(), &proto.UpdatesRequest{
			Metrics: []*proto.Metric{
				{
					Data: &proto.Metric_Delta{Delta: 1},
					Id:   "error",
					Type: proto.Types_GAUGE,
				},
			},
		})

		require.Error(t, err)
	})
}
