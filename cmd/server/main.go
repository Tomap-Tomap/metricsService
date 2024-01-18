package main

import (
	"net/http"

	"github.com/DarkOmap/metricsService/internal/handlers"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/storage"
)

func main() {
	p := parameters.ParseFlagsServer()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	ms := storage.NewMemStorage()
	sh := handlers.NewServiceHandlers(ms)
	r := handlers.ServiceRouter(sh)

	err := http.ListenAndServe(p.FlagRunAddr, r)

	if err != nil {
		panic(err)
	}
}
