package main

import (
	"net/http"

	"github.com/DarkOmap/metricsService/internal/handlers"
)

func main() {
	r := handlers.ServiceRouter()

	err := http.ListenAndServe(":8080", r)

	if err != nil {
		panic(err)
	}
}
