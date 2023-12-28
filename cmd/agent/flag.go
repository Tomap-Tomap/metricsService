package main

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var reportInterval, pollInterval uint

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to server")
	flag.UintVar(&reportInterval, "r", 10, "report interval")
	flag.UintVar(&pollInterval, "p", 2, "poll interval")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		flagRunAddr = envAddr
	}

	if envRI := os.Getenv("REPORT_INTERVAL"); envRI != "" {
		intRI, err := strconv.ParseUint(envRI, 10, 32)

		if err == nil {
			reportInterval = uint(intRI)
		}
	}

	if envPI := os.Getenv("POLL_INTERVAL"); envPI != "" {
		intPI, err := strconv.ParseUint(envPI, 10, 32)

		if err == nil {
			pollInterval = uint(intPI)
		}
	}
}
