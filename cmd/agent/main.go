package main

import (
	"github.com/DarkOmap/metricsService/internal/agent"
	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/DarkOmap/metricsService/internal/parameters"
)

func main() {
	p := parameters.ParseFlagsAgent()
	c := client.NewClient(p.ListenAddr)
	a := agent.NewAgent(c, p.ReportInterval, p.PollInterval)
	a.Run()
}
