// Agent main package.
// Agent collects metrics and sends them to the server
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/DarkOmap/metricsService/internal/agent"
	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"go.uber.org/zap"
)

func main() {
	p := parameters.ParseFlagsAgent()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	logger.Log.Info("Create client")
	c := client.NewClient(p.ListenAddr, p.Key, p.RateLimit)
	defer c.Close()
	logger.Log.Info("Create agent")
	a, err := agent.NewAgent(c, p.ReportInterval, p.PollInterval)

	if err != nil {
		logger.Log.Fatal("Create agent", zap.Error(err))
	}

	logger.Log.Info("Agent start")

	go func() {
		log.Println(http.ListenAndServe("localhost:8082", nil))
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	// defer cancel()
	err = a.Run(ctx)
	if err != nil {
		logger.Log.Fatal("Run agent", zap.Error(err))
	}

	// fmem, err := os.Create(`result.pprof`)
	// if err != nil {
	// 	panic(err)
	// }
	// defer fmem.Close()
	// runtime.GC() // получаем статистику по использованию памяти
	// if err := pprof.WriteHeapProfile(fmem); err != nil {
	// 	panic(err)
	// }
}
