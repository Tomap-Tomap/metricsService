package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
)

type Repository interface {
	UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	GetAllGauge(ctx context.Context) (map[string]storage.Gauge, error)
	GetAllCounter(ctx context.Context) (map[string]storage.Counter, error)
	PingDB(ctx context.Context) error
	Updates(ctx context.Context, metrics []models.Metrics) error
}

type ServiceHandlers struct {
	ms  Repository
	key string
}

func (sh *ServiceHandlers) updateByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Type", "charset=utf-8")

	m, err := getModelsByJSON(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err = sh.ms.UpdateByMetrics(r.Context(), *m)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(m)

	if err != nil {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = sh.ms.UpdateByMetrics(r.Context(), *m)

	if err != nil {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err = sh.ms.ValueByMetrics(r.Context(), *m)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(m)

	if err != nil {
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	m, err = sh.ms.ValueByMetrics(r.Context(), *m)

	if err != nil {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type resultTable struct {
		Counters map[string]storage.Counter
		Gauges   map[string]storage.Gauge
	}

	ctx := r.Context()

	counters, err := sh.ms.GetAllCounter(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gauges, err := sh.ms.GetAllGauge(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := resultTable{counters, gauges}
	err = t.Execute(w, result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (sh *ServiceHandlers) ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	err := sh.ms.PingDB(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (sh *ServiceHandlers) updates(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := models.GetModelsSliceByJSON(buf.Bytes())

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = sh.ms.Updates(r.Context(), m)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func getModelsByJSON(body io.ReadCloser) (*models.Metrics, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)

	if err != nil {
		return nil, err
	}

	m, err := models.NewModelsByJSON(buf.Bytes())

	if err != nil {
		return nil, err
	}

	return m, err
}

func NewServiceHandlers(ms Repository, key string) ServiceHandlers {
	return ServiceHandlers{ms, key}
}

func ServiceRouter(sh ServiceHandlers) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(sh.requestHash)
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
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", sh.ping)
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", sh.updates)
		})
	})

	return r
}

type (
	hashingResponseWriter struct {
		http.ResponseWriter
		bytes int
		key   string
	}
)

func (r *hashingResponseWriter) Write(b []byte) (int, error) {
	h := hmac.New(sha256.New, []byte(r.key))
	h.Write(b)
	dst := h.Sum(nil)

	r.ResponseWriter.Header().Add("HashSHA256", hex.EncodeToString(dst))

	size, err := r.ResponseWriter.Write(b)
	r.bytes += size
	return size, err
}

func (sh *ServiceHandlers) requestHash(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		if sh.key == "" {
			h.ServeHTTP(w, r)
			return
		}

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(&buf)
		hash := hmac.New(sha256.New, []byte(sh.key))
		hash.Write(buf.Bytes())
		dst := hash.Sum(nil)
		hh, err := hex.DecodeString(r.Header.Get("HashSHA256"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !hmac.Equal(hh, dst) {
			http.Error(w, "hash not equal", http.StatusBadRequest)
			return
		}

		hw := hashingResponseWriter{
			ResponseWriter: w,
			key:            sh.key,
		}

		h.ServeHTTP(&hw, r)
	}

	return http.HandlerFunc(logFn)
}
