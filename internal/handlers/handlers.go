package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/go-chi/chi/v5"
)

type repository interface {
	SetGauge(value string, name string) error
	GetGauge(name string) (storage.Gauge, error)
	AddCounter(value string, name string) error
	GetCounter(name string) (storage.Counter, error)
	GetAllGauge() map[string]storage.Gauge
	GetAllCounter() map[string]storage.Counter
}

type ServiceHandlers struct {
	ms repository
}

func (sh *ServiceHandlers) updateCounter(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Content-Type", "charset=utf-8")

	t := strings.ToLower(chi.URLParam(req, "type"))
	v, n := chi.URLParam(req, "value"), chi.URLParam(req, "name")
	switch t {
	case "counter":
		err := sh.ms.AddCounter(v, n)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case "gauge":
		err := sh.ms.SetGauge(v, n)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(res, "unknown type", http.StatusBadRequest)
	}
}

func (sh *ServiceHandlers) value(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Content-Type", "charset=utf-8")

	t := strings.ToLower(chi.URLParam(req, "type"))

	switch t {
	case "counter":
		v, err := sh.ms.GetCounter(chi.URLParam(req, "name"))

		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}

		fmt.Fprint(res, v)
	case "gauge":
		v, err := sh.ms.GetGauge(chi.URLParam(req, "name"))

		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprint(res, v)
	default:
		http.Error(res, "unknown type", http.StatusNotFound)
	}
}

func (sh *ServiceHandlers) all(res http.ResponseWriter, req *http.Request) {
	tmpl := `<!DOCTYPE html>
	<html>
	
	<head>
		<meta charset="UTF-8">
		<title>add data from service</title>
	</head>
	
	<body>
		<table>
			<tr>
				<th>name</th>
				<th>value</th>
			</tr>
			{{ range $s, $v := .Counters }}
			<tr>
				<td>{{ $s }}</td>
				<td>{{ $v }}</td>
			</tr>
			{{end}}
			{{ range $s, $v := .Gauges }}
			<tr>
				<td>{{ $s }}</td>
				<td>{{ $v }}</td>
			</tr>
			{{end}}
		</table>
	</body>
	
	</html>`

	res.Header().Add("Content-Type", "text/html")
	res.Header().Add("Content-Type", "charset=utf-8")

	t := template.New("all tmpl")
	t, err := t.Parse(tmpl)

	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	type resultTable struct {
		Counters map[string]storage.Counter
		Gauges   map[string]storage.Gauge
	}

	result := resultTable{sh.ms.GetAllCounter(), sh.ms.GetAllGauge()}
	err = t.Execute(res, result)

	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
}

func NewServiceHandlers(ms repository) ServiceHandlers {
	return ServiceHandlers{ms}
}

func ServiceRouter(sh ServiceHandlers) chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", sh.all)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", sh.updateCounter)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", sh.value)
		})
	})

	return r
}
