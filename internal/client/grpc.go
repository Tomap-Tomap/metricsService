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

// GRPC client's grpc structure
type GRPC struct {
	client proto.MetricsClient
	hasher *hasher.Hasher
	conn   *grpc.ClientConn
}

// NewGRPC create new grpc client
func NewGRPC(p parameters.AgentParameters) (*GRPC, error) {
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

	return &GRPC{
		client: proto.NewMetricsClient(conn),
		hasher: h,
		conn:   conn,
	}, nil
}

// Close closes grpc client
func (gc *GRPC) Close() error {
	gc.hasher.Close()
	err := gc.conn.Close()
	if err != nil {
		return fmt.Errorf("close connetion: %w", err)
	}

	return nil
}

// SendGauge sends gauge metric to server
func (gc *GRPC) SendGauge(ctx context.Context, name string, value float64) error {
	metric := proto.Metric{
		Data: &proto.Metric_Value{Value: value},
		Id:   name,
		Type: proto.Types_GAUGE,
	}

	_, err := gc.client.Update(ctx, &proto.UpdateRequest{Metric: &metric})
	if err != nil {
		return fmt.Errorf("send gauge metric with grpc name %s value %f: %w", name, value, err)
	}

	return nil
}

// SendCounter sends counter metric to server
func (gc *GRPC) SendCounter(ctx context.Context, name string, delta int64) error {
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

// SendBatch sends metrics to server
func (gc *GRPC) SendBatch(ctx context.Context, batch map[string]float64) error {
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
		return fmt.Errorf("send batch in grpc: %w", err)
	}

	return nil
}
