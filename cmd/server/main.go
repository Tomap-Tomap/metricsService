package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

var ms MemStorage

func (m MemStorage) addGauges(name string, g gauge) error {
	m.gauges[name] = g

	return nil
}

func (m MemStorage) addCounter(name string, c counter) error {
	m.counters[name] += c

	return nil
}

func update(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseForm()

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	url := req.URL.RequestURI()
	param := strings.Split(url, "/")

	if len(param) < 3 {
		http.Error(res, "URI path is to short", http.StatusNotFound)
		return
	}

	if param[1] == "" {
		http.Error(res, "Metrics name is empty", http.StatusNotFound)
		return
	}

	switch param[0] {
	case "gauge":
		g, err := strconv.ParseFloat(param[2], 64)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		ms.addGauges(param[1], gauge(g))
	case "counter":
		c, err := strconv.ParseInt(param[2], 10, 64)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		ms.addCounter(param[1], counter(c))
	default:
		http.Error(res, "Metrics type is unknown", http.StatusBadRequest)
	}
}

func main() {
	ms.counters = make(map[string]counter)
	ms.gauges = make(map[string]gauge)

	updatePath := "/update/"
	mux := http.NewServeMux()
	mux.Handle(updatePath, http.StripPrefix(updatePath, http.HandlerFunc(update)))

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}
