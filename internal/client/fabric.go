package client

import (
	"context"
	"fmt"

	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
)

// Client describes client methods
type Client interface {
	SendBatch(ctx context.Context, batch map[string]float64) error
	SendCounter(ctx context.Context, name string, delta int64) error
	SendGauge(ctx context.Context, name string, value float64) error
	Close() error
}

// NewClient fabric method to create client
func NewClient(p parameters.AgentParameters) (Client, error) {
	if p.UseGRPC {
		logger.Log.Info("Create grpc client")
		c, err := NewGRPC(p)
		if err != nil {
			return nil, fmt.Errorf("create grpc client: %w", err)
		}
		return c, nil
	}

	logger.Log.Info("Create http client")
	c, err := NewHTTP(p)
	if err != nil {
		return nil, fmt.Errorf("create http client: %w", err)
	}

	return c, nil
}
