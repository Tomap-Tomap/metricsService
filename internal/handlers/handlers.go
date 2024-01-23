package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type repository interface {
	UpdateByMetrics(m models.Metrics) (models.Metrics, error)
	ValueByMetrics(m models.Metrics) (models.Metrics, error)
	GetAllGauge() map[string]storage.Gauge
	GetAllCounter() map[string]storage.Counter
}

type ServiceHandlers struct {
	ms repository
}

func (sh *ServiceHandlers) updateByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Type", "charset=utf-8")

	m, err := getModelsByJSON(r.Body)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "updateByJSON"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err = sh.ms.UpdateByMetrics(m)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "updateByJSON"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(m)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "updateByJSON"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (sh *ServiceHandlers) updateByURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	m, err := models.NewModelByStrings(
		chi.URLParam(r, "name"),
		chi.URLParam(r, "type"),
		chi.URLParam(r, "value"),
	)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "updateByURL"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = sh.ms.UpdateByMetrics(m)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "updateByURL"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (sh *ServiceHandlers) valueByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Type", "charset=utf-8")

	m, err := getModelsByJSON(r.Body)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "valueByJSON"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err = sh.ms.ValueByMetrics(m)

	if err != nil {
		logger.Log.Info("value not found",
			zap.String("handler", "valueByJSON"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(m)

	if err != nil {
		logger.Log.Info("got incorrect request",
			zap.String("handler", "valueByJSON"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (sh *ServiceHandlers) valueByURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	m, err := models.NewMetrics(
		chi.URLParam(r, "name"),
		chi.URLParam(r, "type"),
	)

	if err != nil {
		logger.Log.Info("value not found",
			zap.String("handler", "valueByURL"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	m, err = sh.ms.ValueByMetrics(m)

	if err != nil {
		logger.Log.Info("value not found",
			zap.String("handler", "valueByURL"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if m.Delta != nil {
		fmt.Fprint(w, *m.Delta)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, *m.Value)
}

func (sh *ServiceHandlers) all(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Content-Type", "charset=utf-8")

	t := template.New("all tmpl")
	t, err := t.Parse(tmpl)

	if err != nil {
		logger.Log.Info("html parse",
			zap.String("handler", "all"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), 500)
		return
	}

	type resultTable struct {
		Counters map[string]storage.Counter
		Gauges   map[string]storage.Gauge
	}

	result := resultTable{sh.ms.GetAllCounter(), sh.ms.GetAllGauge()}
	err = t.Execute(w, result)

	if err != nil {
		logger.Log.Info("html execute",
			zap.String("handler", "all"),
			zap.String("uri", r.RequestURI),
			zap.Error(err),
		)
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getModelsByJSON(body io.ReadCloser) (models.Metrics, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)

	if err != nil {
		return models.Metrics{}, err
	}

	m, err := models.NewModelsByJSON(buf.Bytes())

	if err != nil {
		return models.Metrics{}, err
	}

	return m, err
}

func NewServiceHandlers(ms repository) ServiceHandlers {
	return ServiceHandlers{ms}
}

func ServiceRouter(sh ServiceHandlers) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compresses.CompressHandle)

	r.Route("/", func(r chi.Router) {
		r.Get("/", sh.all)
		r.Route("/update", func(r chi.Router) {
			r.Post("/", sh.updateByJSON)
			r.Post("/{type}/{name}/{value}", sh.updateByURL)
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", sh.valueByJSON)
			r.Get("/{type}/{name}", sh.valueByURL)
		})
	})

	return r
}
