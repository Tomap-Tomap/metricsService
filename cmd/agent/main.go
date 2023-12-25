package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/DarkOmap/metricsService/internal/client"
	memstats "github.com/DarkOmap/metricsService/internal/memStats"
)

func main() {
	timeToPush := byte(0)
	sleepTime := byte(2)
	pollCount := 0

	for {
		if timeToPush == 10 {
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

		time.Sleep(time.Duration(sleepTime) * time.Second)
		timeToPush += sleepTime
		pollCount++
	}
}
