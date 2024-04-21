// The package agent defines a structure that gets memory metrics and sends them to the server.
package agent

import (
	"context"
	"fmt"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/memstats"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Client it's type for sending data to server.
type Client interface {
	SendBatch(ctx context.Context, batch map[string]float64) error
	SendCounter(ctx context.Context, name string, delta int64) error
}

// Agent it's structure for calculate and send data to server.
type Agent struct {
	reportInterval uint
	pollInterval   uint
	rateLimit      uint
	client         Client
	pollCount      atomic.Int64
	ms             *memstats.MemStatsForServer
}

func NewAgent(client Client, reportInterval, pollInterval, rateLimit uint) (*Agent, error) {
	a := &Agent{reportInterval: reportInterval,
		pollInterval: pollInterval,
		rateLimit:    rateLimit,
		client:       client}

	ms, err := memstats.NewMemStatsForServer()

	if err != nil {
		return nil, fmt.Errorf("create mem stats")
	}

	a.ms = ms

	return a, nil
}

// Run start calculate and sending data to server.
func (a *Agent) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	eg, egCtx := errgroup.WithContext(ctx)

	jobs := make(chan func(context.Context) error, a.rateLimit)
	defer close(jobs)
	for w := 1; w <= cap(jobs); w++ {
		go worker(egCtx, jobs)
	}

	eg.Go(func() error {
		err := a.startSendReport(egCtx, jobs)
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

func (a *Agent) startSendReport(ctx context.Context, jobs chan<- func(context.Context) error) error {
	logger.Log.Info("Send report start")
	for {
		select {
		case <-time.After(time.Duration(a.reportInterval) * time.Second):
			jobs <- a.sendBatch
			jobs <- a.sendCounter
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

func (a *Agent) sendBatch(ctx context.Context) error {
	msForServer := a.ms.GetMap()

	return a.client.SendBatch(ctx, msForServer)
}

func (a *Agent) sendCounter(ctx context.Context) error {
	return a.client.SendCounter(ctx, "PollCount", a.pollCount.Load())
}

func worker(ctx context.Context, jobs <-chan func(context.Context) error) {
	for j := range jobs {
		err := j(ctx)

		if err != nil {
			logger.Log.Warn(
				"Error on sending to server",
				zap.Error(err),
			)
		}
	}
}
