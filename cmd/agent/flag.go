package main

import (
	"flag"
)

var flagRunAddr string
var reportInterval, pollInterval uint

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "http://localhost:8080", "address and port to server")
	flag.UintVar(&reportInterval, "r", 10, "report interval")
	flag.UintVar(&pollInterval, "p", 2, "poll interval")
	flag.Parse()
}
