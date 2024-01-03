package main

import (
	"net/http"

	"github.com/DarkOmap/metricsService/internal/handlers"
	"github.com/DarkOmap/metricsService/internal/parameters"
)

func main() {
	flagRunAddr := parameters.ParseFlagsServer()

	r := handlers.ServiceRouter()

	err := http.ListenAndServe(flagRunAddr, r)

	if err != nil {
		panic(err)
	}
}
