// Agent main package.
// Agent collects metrics and sends them to the server
package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/DarkOmap/metricsService/internal/agent"
	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/memstats"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"go.uber.org/zap"
)

func main() {
	p := parameters.ParseFlagsAgent()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	logger.Log.Info("Create gzip pool")
	pool := compresses.NewGzipPool(p.RateLimit)
	defer pool.Close()
	logger.Log.Info("Create hasher pool")
	h := hasher.NewHasher([]byte(p.Key), p.RateLimit)
	defer h.Close()
	logger.Log.Info("Create client")
	c := client.NewClient(pool, h, p.ListenAddr)
	logger.Log.Info("Init mem stats")
	ms, err := memstats.NewMemStatsForServer()

	if err != nil {
		logger.Log.Fatal("create mem stats", zap.Error(err))
	}

	logger.Log.Info("Create agent")
	a := agent.NewAgent(c, ms, p.ReportInterval, p.PollInterval)

	logger.Log.Info("Agent start")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = a.Run(ctx)
	if err != nil {
		logger.Log.Fatal("Run agent", zap.Error(err))
	}
}
