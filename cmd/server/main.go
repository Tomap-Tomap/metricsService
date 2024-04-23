// Server main package.
// Server defines handlers for collecting metrics and stores them in the database.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/DarkOmap/metricsService/internal/file"
	"github.com/DarkOmap/metricsService/internal/handlers"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/storage"
	_ "github.com/DarkOmap/metricsService/swagger"
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
	p := parameters.ParseFlagsServer()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	logger.Log.Info("Create file producer")
	producer, err := file.NewProducer(p.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Create file producer", zap.Error(err))
	}
	defer producer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	eg, egCtx := errgroup.WithContext(ctx)

	var ms handlers.Repository

	if p.DataBaseDSN != "" {
		logger.Log.Info("Create database storage")

		conn, err := pgxpool.New(ctx, p.DataBaseDSN)

		if err != nil {
			logger.Log.Fatal("Connect to database", zap.Error(err))
		}
		defer conn.Close()

		ms, err = storage.NewDBStorage(conn)

		if err != nil {
			logger.Log.Fatal("Create database storage", zap.Error(err))
		}
	} else {
		logger.Log.Info("Create mem storage")
		ms, err = storage.NewMemStorage(egCtx, eg, producer, p)
		if err != nil {
			logger.Log.Fatal("Create mem storage", zap.Error(err))
		}
	}

	logger.Log.Info("Create handlers")
	sh := handlers.NewServiceHandlers(ms)
	logger.Log.Info("Create routers")
	r := handlers.ServiceRouter(sh, p.Key)

	logger.Log.Info("Create server")
	httpServer := &http.Server{
		Addr:    p.FlagRunAddr,
		Handler: r,
	}

	eg.Go(func() error {
		logger.Log.Info("Run serve")
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-egCtx.Done()
		logger.Log.Info("Stop serve")
		return httpServer.Shutdown(context.Background())
	})

	if err := eg.Wait(); err != nil {
		logger.Log.Fatal("Problem with working server", zap.Error(err))
	}
}
