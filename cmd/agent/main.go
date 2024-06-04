// Agent main package.
// Agent collects metrics and sends them to the server
package main

import (
	"cmp"
	"context"
	"fmt"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/DarkOmap/metricsService/internal/agent"
	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/memstats"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	displayBuild(buildVersion, buildDate, buildCommit)
	p := parameters.ParseFlagsAgent()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	logger.Log.Info("Create client")
	c, err := client.NewClient(p)

	if err != nil {
		logger.Log.Fatal("Create client", zap.Error(err))
	}
	defer c.Close()

	logger.Log.Info("Init mem stats")
	ms, err := memstats.NewMemStatsForServer()

	if err != nil {
		logger.Log.Fatal("Create mem stats", zap.Error(err))
	}

	logger.Log.Info("Create agent")
	a := agent.NewAgent(c, ms, p.ReportInterval, p.PollInterval)

	logger.Log.Info("Agent start")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	err = a.Run(ctx)
	if err != nil {
		logger.Log.Fatal("Run agent", zap.Error(err))
	}
}

func displayBuild(version, date, commit string) (string, string, string) {
	version = cmp.Or(version, "N/A")
	date = cmp.Or(date, "N/A")
	commit = cmp.Or(commit, "N/A")

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)

	return version, date, commit
}
