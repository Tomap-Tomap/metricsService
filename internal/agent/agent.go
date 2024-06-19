// Package agent defines a structure that gets memory metrics and sends them to the server.
package agent

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/DarkOmap/metricsService/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Client it's type for sending data to server.
type Client interface {
	SendBatch(ctx context.Context, batch map[string]float64) error
	SendCounter(ctx context.Context, name string, delta int64) error
}

// MemStats it's type for reading memory statistics
type MemStats interface {
	ReadMemStats() error
	GetMap() map[string]float64
}

// Agent it's structure for calculate and send data to server.
type Agent struct {
	client         Client
	ms             MemStats
	pollCount      atomic.Int64
	reportInterval uint
	pollInterval   uint
}

// NewAgent create agent
func NewAgent(client Client, ms MemStats, reportInterval, pollInterval uint) *Agent {
	a := &Agent{
		reportInterval: reportInterval,
		pollInterval:   pollInterval,
		client:         client,
		ms:             ms,
	}

	return a
}

// Run start calculate and sending data to server.
func (a *Agent) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		a.startSendReport(egCtx)
		return nil
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

func (a *Agent) startSendReport(ctx context.Context) {
	logger.Log.Info("Send report start")
	for {
		select {
		case <-time.After(time.Duration(a.reportInterval) * time.Second):
			a.sendMemStats(ctx)
			a.sendPollCount(ctx)
		case <-ctx.Done():
			logger.Log.Info("Send report done")
			return
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

func (a *Agent) sendMemStats(ctx context.Context) {
	msForServer := a.ms.GetMap()

	err := a.client.SendBatch(ctx, msForServer)
	if err != nil {
		logger.Log.Warn("Send batch", zap.Error(err))
	}
}

func (a *Agent) sendPollCount(ctx context.Context) {
	err := a.client.SendCounter(ctx, "PollCount", a.pollCount.Load())
	if err != nil {
		logger.Log.Warn("Send counter", zap.Error(err))
	}
}
