// Server main package.
// Server defines handlers for collecting metrics and stores them in the database.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/DarkOmap/metricsService/internal/build"
	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/ip"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/server"
	_ "github.com/DarkOmap/metricsService/swagger"
	_ "google.golang.org/grpc/encoding/gzip"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

//	@Title			MetricsSevice API
//	@Description	Service to communicate with storage.
//	@Version		1.0

//	@Contact.email	timur.konoplev@yandex.ru

//	@BasePath	/
//	@Host		localhost:8080

//	@SecurityDefinitions.apikey	ApiKeyAuth
//	@In							header
//	@Name						HashSHA256

//	@Tag.name			Update
//	@Tag.description	"Query group for updates on metrics data"

//	@Tag.name			Value
//	@Tag.description	"Query group for metrics data retrieval"

func main() {
	build.DisplayBuild(buildVersion, buildDate, buildCommit)
	p := parameters.ParseFlagsServer()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	logger.Log.Info("Create repository")
	r, err := server.NewRepository(ctx, p)
	if err != nil {
		logger.Log.Fatal("Create repository", zap.Error(err))
	}

	defer func() {
		err := r.Close()

		logger.Log.Fatal("Close repository", zap.Error(err))
	}()

	logger.Log.Info("Create IP checker")
	ipc := ip.NewChecker(p.TrustedSubnet)

	logger.Log.Info("Create hasher pool")
	h := hasher.NewHasher([]byte(p.HashKey), p.RateLimit)
	defer h.Close()

	logger.Log.Info("Create gzip pool")
	gzipPool := compresses.NewGzipPool(p.RateLimit)
	defer gzipPool.Close()

	opts := make([]server.OptionFunc, 0, 2)

	if p.FlagRunAddr != "" {
		opts = append(opts, server.WithHTTP(r, ipc, h, gzipPool, p))
	}

	if p.FlagRunGRPCAddr != "" {
		opts = append(opts, server.WithGRPC(r, ipc, h, p))
	}

	logger.Log.Info("Create server")
	server, err := server.NewServer(opts...)
	if err != nil {
		logger.Log.Fatal("Create server", zap.Error(err))
	}

	if err := server.Run(ctx); err != nil {
		logger.Log.Fatal("Problem with working server", zap.Error(err))
	}
}
