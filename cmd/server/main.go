package main

import (
	"net/http"

	"github.com/DarkOmap/metricsService/internal/handlers"
)

func main() {
	parseFlags()

	r := handlers.ServiceRouter()

	err := http.ListenAndServe(flagRunAddr, r)

	if err != nil {
		panic(err)
	}
}
