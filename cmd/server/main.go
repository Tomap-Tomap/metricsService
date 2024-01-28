package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/DarkOmap/metricsService/internal/handlers"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/storage"
	"go.uber.org/zap"
)

func main() {
	var (
		err error
		wg  sync.WaitGroup
	)

	p := parameters.ParseFlagsServer()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	ms := &storage.MemStorage{}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if p.Restore {
		ms, err = storage.NewMemStorageFromGile(ctx, &wg, p.StoreInterval, p.FileStoragePath)

		if err != nil {
			logger.Log.Fatal("create memory storage", zap.Error(err))
		}
	} else {
		ms = storage.NewMemStorage(ctx, &wg, p.StoreInterval, p.FileStoragePath)
	}

	sh := handlers.NewServiceHandlers(ms)
	r := handlers.ServiceRouter(sh)

	srv := &http.Server{
		Addr:    p.FlagRunAddr,
		Handler: r,
	}

	var wgDone sync.WaitGroup
	wgDone.Add(2)
	go func() error {
		defer wgDone.Done()
		fmt.Println("start")
		return srv.ListenAndServe()
	}()

	go func() error {
		defer wgDone.Done()
		wg.Wait()
		<-ctx.Done()
		fmt.Println("test")
		return srv.Shutdown(ctx)
	}()

	wgDone.Wait()

	fmt.Println("test3")
}
