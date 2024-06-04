package client

import (
	"context"
	"fmt"

	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/ip"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
)

type ClientGRPC struct {
	client proto.MetricsClient
	hasher *hasher.Hasher
	conn   *grpc.ClientConn
}

func NewClientGRPC(p parameters.AgentParameters) (*ClientGRPC, error) {
	logger.Log.Info("Create hasher pool")
	h := hasher.NewHasher([]byte(p.HashKey), p.RateLimit)

	conn, err := grpc.NewClient(
		p.ListenAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			ip.InterceptorAddRealIP,
			h.InterceptorAddHashMD,
		),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)

	if err != nil {
		return nil, fmt.Errorf("create grpc client conntection: %w", err)
	}

	return &ClientGRPC{
		client: proto.NewMetricsClient(conn),
		hasher: h,
		conn:   conn,
	}, nil
}

func (gc *ClientGRPC) Close() {
	gc.hasher.Close()
	gc.conn.Close()
}

func (gc *ClientGRPC) SendGauge(ctx context.Context, name string, value float64) error {
	metric := proto.Metric{
		Data: &proto.Metric_Value{Value: value},
		Id:   name,
		Type: proto.Types_GAUGE,
	}

	_, err := gc.client.Update(ctx, &proto.UpdateRequest{Metric: &metric})

	if err != nil {
		return fmt.Errorf("send gauge: %w", err)
	}

	return nil
}

func (gc *ClientGRPC) SendCounter(ctx context.Context, name string, delta int64) error {
	metric := proto.Metric{
		Data: &proto.Metric_Delta{Delta: delta},
		Id:   name,
		Type: proto.Types_COUNTER,
	}

	_, err := gc.client.Update(ctx, &proto.UpdateRequest{Metric: &metric})

	if err != nil {
		return fmt.Errorf("send conter: %w", err)
	}

	return nil
}

func (gc *ClientGRPC) SendBatch(ctx context.Context, batch map[string]float64) error {
	metrics := make([]*proto.Metric, 0, len(batch))

	for i, v := range batch {
		metrics = append(metrics, &proto.Metric{
			Data: &proto.Metric_Value{Value: v},
			Id:   i,
			Type: proto.Types_GAUGE,
		})
	}

	_, err := gc.client.Updates(ctx, &proto.UpdatesRequest{Metrics: metrics})

	if err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	return nil
}
