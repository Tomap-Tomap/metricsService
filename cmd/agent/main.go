package main

import (
	"github.com/DarkOmap/metricsService/internal/agent"
	"github.com/DarkOmap/metricsService/internal/parameters"
)

func main() {
	listenAddr, reportInterval, pollInterval := parameters.ParseFlagsAgent()

	agent.Run(listenAddr, reportInterval, pollInterval)
}
