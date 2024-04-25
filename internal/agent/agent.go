// The package agent defines a structure that gets memory metrics and sends them to the server.
package agent

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/memstats"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Client it's type for sending data to server.
type Client interface {
	SendBatch(ctx context.Context, batch map[string]float64)
	SendCounter(ctx context.Context, name string, delta int64)
}

// Agent it's structure for calculate and send data to server.
type Agent struct {
	reportInterval uint
	pollInterval   uint
	client         Client
	pollCount      atomic.Int64
	ms             *memstats.MemStatsForServer
}

func NewAgent(client Client, reportInterval, pollInterval uint) (*Agent, error) {
	a := &Agent{reportInterval: reportInterval,
		pollInterval: pollInterval,
		client:       client}

	ms, err := memstats.NewMemStatsForServer()

	if err != nil {
		return nil, fmt.Errorf("create mem stats")
	}

	a.ms = ms

	return a, nil
}

// Run start calculate and sending data to server.
func (a *Agent) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err := a.startSendReport(egCtx)
		return err
	})

	eg.Go(func() error {
		err := a.startReadMemStats(egCtx)
		return err
	})

	if err := eg.Wait(); err != nil {
		logger.Log.Error("Problem with working agent", zap.Error(err))
		return fmt.Errorf("problem with working agent: %w", err)
	}

	return nil
}

func (a *Agent) startSendReport(ctx context.Context) error {
	logger.Log.Info("Send report start")
	for {
		select {
		case <-time.After(time.Duration(a.reportInterval) * time.Second):
			a.sendBatch(ctx)
			a.sendCounter(ctx)
		case <-ctx.Done():
			logger.Log.Info("Send report done")
			return nil
		}
	}
}

func (a *Agent) startReadMemStats(ctx context.Context) error {
	logger.Log.Info("Read mem stats start")
	for {
		select {
		case <-time.After(time.Duration(a.pollInterval) * time.Second):
			err := a.ms.ReadMemStats()

			if err != nil {
				return fmt.Errorf("read mem stats: %w", err)
			}

			a.pollCount.Add(1)
		case <-ctx.Done():
			logger.Log.Info("Read mem stats done")
			return nil
		}
	}
}

func (a *Agent) sendBatch(ctx context.Context) {
	msForServer := a.ms.GetMap()

	a.client.SendBatch(ctx, msForServer)
}

func (a *Agent) sendCounter(ctx context.Context) {
	a.client.SendCounter(ctx, "PollCount", a.pollCount.Load())
}
