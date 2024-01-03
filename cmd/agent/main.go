package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/DarkOmap/metricsService/internal/client"
	memstats "github.com/DarkOmap/metricsService/internal/memStats"
	"github.com/DarkOmap/metricsService/internal/parameters"
)

func main() {
	listenAddr, reportInterval, pollInterval := parameters.ParseFlagsAgent()

	timeToPush := uint(0)
	pollCount := 0

	client.ServiceAddr = listenAddr + "/update"

	for {
		if timeToPush == reportInterval {
			ms := memstats.GetMemStatsForServer()
			err := client.PushStats(ms)

			if err != nil {
				panic(err.Error())
			}

			pollCountString := strconv.Itoa(pollCount)
			err = client.SendCounter("PollCount", pollCountString)

			if err != nil {
				panic(err.Error())
			}

			err = client.SendGauge("RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))

			if err != nil {
				panic(err.Error())
			}

			timeToPush = 0
		}

		memstats.ReadMemStats()

		time.Sleep(time.Duration(pollInterval) * time.Second)
		timeToPush += pollInterval
		pollCount++
	}
}
