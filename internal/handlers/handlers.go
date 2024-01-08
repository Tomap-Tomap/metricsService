package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/go-chi/chi/v5"
)

type ServiceHandlers struct {
	ms storage.Repositories
}

func (sh *ServiceHandlers) updateCounter(res http.ResponseWriter, req *http.Request) {
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

	err = sh.ms.AddValue(v, chi.URLParam(req, "name"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (sh *ServiceHandlers) value(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Content-Type", "charset=utf-8")

	t, err := storage.ParseType(chi.URLParam(req, "type"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	v, err := sh.ms.GetValue(t, chi.URLParam(req, "name"))

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprint(res, v)
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
			{{ range $s, $v := . }}
			<tr>
				<td>{{ $v.Name }}</td>
				<td>{{ $v.Value }}</td>
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

	tableResult := sh.ms.GetData()
	err = t.Execute(res, tableResult)

	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
}

func NewServiceHandlers(ms storage.Repositories) ServiceHandlers {
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
