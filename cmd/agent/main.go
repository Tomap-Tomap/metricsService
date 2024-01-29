package main

import (
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

	logger.Log.Info("create client")
	c := client.NewClient(p.ListenAddr)
	logger.Log.Info("create agent")
	a := agent.NewAgent(c, p.ReportInterval, p.PollInterval)
	logger.Log.Info("agent start")
	err := a.Run()
	if err != nil {
		logger.Log.Fatal("run agent", zap.Error(err))
	}
}
