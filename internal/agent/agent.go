package agent

import (
	"context"
	"math/rand"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/memstats"
	"go.uber.org/zap"
)

type Agent struct {
	reportInterval, pollInterval uint
	client                       client.Client
	pollCount                    atomic.Int64
	ms                           runtime.MemStats
}

func (a *Agent) Run() {
	var wg sync.WaitGroup

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wg.Add(2)

	go func() {
		defer wg.Done()
		loop := true
		for loop {
			select {
			case <-time.After(time.Duration(a.reportInterval) * time.Second):
				a.sendReport(ctx)
			case <-ctx.Done():
				loop = false
			}
		}
	}()

	go func() {
		defer wg.Done()
		loop := true
		for loop {
			select {
			case <-time.After(time.Duration(a.pollInterval) * time.Second):
				runtime.ReadMemStats(&a.ms)
				a.pollCount.Add(1)
			case <-ctx.Done():
				loop = false
			}
		}
	}()

	wg.Wait()
}

func (a *Agent) sendReport(ctx context.Context) {
	msForServer := memstats.GetMemStatsForServer(&a.ms)

	for k, v := range msForServer {
		err := a.client.SendGauge(ctx, k, v)

		if err != nil {
			logger.Log.Warn(
				"push memstats",
				zap.String("name", k),
				zap.Float64("value", v),
				zap.Error(err),
			)
		}
	}

	err := a.client.SendCounter(ctx, "PollCount", a.pollCount.Load())

	if err != nil {
		logger.Log.Warn(
			"Error on sending poll count",
			zap.Int64("value", a.pollCount.Load()),
			zap.Error(err),
		)
	}

	randV := rand.Float64()
	err = a.client.SendGauge(ctx, "RandomValue", randV)

	if err != nil {
		logger.Log.Warn(
			"Error on sending random value",
			zap.Float64("value", randV),
			zap.Error(err),
		)
	}
}

func NewAgent(client client.Client, reportInterval, pollInterval uint) (a *Agent) {
	a = &Agent{reportInterval: reportInterval, pollInterval: pollInterval, client: client}
	runtime.ReadMemStats(&a.ms)
	return
}
