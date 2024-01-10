package agent

import (
	"context"
	"log"
	"math/rand"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/memstats"
)

func Run(listenAddr string, reportInterval, pollInterval uint) {
	var (
		ms    runtime.MemStats
		wg    sync.WaitGroup
		mutex sync.Mutex
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wg.Add(2)

	pollCount := 0

	go func() {
		defer wg.Done()
		loop := true
		for loop {
			select {
			case <-time.After(time.Duration(reportInterval) * time.Second):
				sendReport(&ms, ctx, listenAddr, pollCount, &mutex)
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
			case <-time.After(time.Duration(pollInterval) * time.Second):
				runtime.ReadMemStats(&ms)
				mutex.Lock()
				pollCount++
				mutex.Unlock()
			case <-ctx.Done():
				loop = false
			}
		}
	}()

	wg.Wait()
}

func sendReport(ms *runtime.MemStats, ctx context.Context, listenAddr string, pollCount int, mutex *sync.Mutex) {
	msForServer := memstats.GetMemStatsForServer(ms)
	err := client.PushStats(ctx, listenAddr, msForServer)

	if err != nil {
		log.Print(err.Error())
	}

	mutex.Lock()
	pollCountString := strconv.Itoa(pollCount)
	err = client.SendCounter(ctx, listenAddr, "PollCount", pollCountString)
	mutex.Unlock()

	if err != nil {
		log.Print(err.Error())
	}

	err = client.SendGauge(ctx, listenAddr, "RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))

	if err != nil {
		log.Print(err.Error())
	}
}
