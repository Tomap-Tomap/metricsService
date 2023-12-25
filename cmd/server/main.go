package main

import (
	"net/http"

	"github.com/DarkOmap/metricsService/internal/handlers"
)

const (
	updatePath = "/update/"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle(updatePath, http.StripPrefix(updatePath, http.HandlerFunc(handlers.Update)))

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}
