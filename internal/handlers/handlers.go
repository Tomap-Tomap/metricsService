package handlers

import (
	"fmt"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/go-chi/chi/v5"
)

var ms storage.MemStorage

func init() {
	ms = storage.NewMemStorage()
}

func updateCounter(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Content-Type", "charset=utf-8")

	t, err := storage.ParseType(chi.URLParam(req, "type"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var v storage.Typer

	if t == storage.CounterType {
		v, err = storage.ParseCounter(chi.URLParam(req, "value"))
	} else if t == storage.GaugeType {
		v, err = storage.ParseGauge(chi.URLParam(req, "value"))
	}

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = ms.AddValue(v, chi.URLParam(req, "name"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func value(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Content-Type", "charset=utf-8")

	t, err := storage.ParseType(chi.URLParam(req, "type"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	v, err := ms.GetValue(t, chi.URLParam(req, "name"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprint(res, v)
}

func all(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/html")
	res.Header().Add("Content-Type", "charset=utf-8")

	htmlText := `<!DOCTYPE html>
	<html>
	<head>
	<meta charset="UTF-8">
	<title>add data from service</title>
	</head>
	<body>
	<table>
	<tr><th>name</th><th>value</th></tr>`

	tableTemplate := "<tr><td>%s</td><td>%v</td></tr>"

	tableResult := ms.GetData()

	tableHTML := ""

	for _, val := range tableResult {
		tableHTML += fmt.Sprintf(tableTemplate, val.Name, val.Value)
	}

	htmlText += tableHTML + "</table></body></html>"

	fmt.Fprint(res, htmlText)
}

func ServiceRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", all)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", updateCounter)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", value)
		})
	})

	return r
}
