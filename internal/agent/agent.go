package agent

import (
	"context"
	"log"
	"math/rand"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/memstats"
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
	err := a.client.PushStats(ctx, msForServer)

	if err != nil {
		log.Printf("Error on sending memory stats: %s", err)
	}

	pollCountString := strconv.FormatInt(a.pollCount.Load(), 10)
	err = a.client.SendCounter(ctx, "PollCount", pollCountString)

	if err != nil {
		log.Printf("Error on sending poll count: %s", err)
	}

	err = a.client.SendGauge(ctx, "RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))

	if err != nil {
		log.Printf("Error on sending random value: %s", err)
	}
}

func NewAgent(client client.Client, reportInterval, pollInterval uint) (a *Agent) {
	a = &Agent{reportInterval: reportInterval, pollInterval: pollInterval, client: client}
	runtime.ReadMemStats(&a.ms)
	return
}
