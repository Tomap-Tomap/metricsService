package main

import (
	"github.com/DarkOmap/metricsService/internal/handlers"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/server"
	"github.com/DarkOmap/metricsService/internal/storage"
	"go.uber.org/zap"
)

func main() {
	p := parameters.ParseFlagsServer()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	logger.Log.Info("create mem storage")
	ms := storage.NewMemStorage()
	logger.Log.Info("create handlers")
	sh := handlers.NewServiceHandlers(ms)
	logger.Log.Info("create routers")
	r := handlers.ServiceRouter(sh)
	logger.Log.Info("create server")
	s, err := server.NewServer(ms, p.FileStoragePath, p.StoreInterval, p.Restore)

	if err != nil {
		logger.Log.Fatal("create server", zap.Error(err))
	}

	logger.Log.Info("server run")
	if err := s.ListenAndServe(p.FlagRunAddr, r); err != nil {
		logger.Log.Fatal("server run", zap.Error(err))
	}
}
