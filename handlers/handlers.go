// Package handlers contain handlers for the server.
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	contentTypeTextPlain       = "text/plain"
	contentTypeApplicationJSON = "application/json"
	contetntTypeTextHTML       = "text/html"

	contentTypeCharsetUTF8 = "charset=utf-8"

	headerContentType = "Content-Type"
)

// Decrypter describes the type for decrypting messages
type Decrypter interface {
	RequestDecrypt(next http.Handler) http.Handler
}

// IPChecker describes the type for checking the IP address
type IPChecker interface {
	RequsetIPCheck(next http.Handler) http.Handler
}

// ServiceHandlers structure with handlers
type ServiceHandlers struct {
	ms Repository
}

// NewServiceHandlers create ServiceHandlers
func NewServiceHandlers(ms Repository) ServiceHandlers {
	return ServiceHandlers{ms}
}

// UpdateByJSON godoc
//
//	@Tags			Update
//	@Summary		Update metrics data
//	@Description	Create new or update existing metric data.
//	@ID				updateUpdateByJSON
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.Metrics	true	"A JSON object with `id`, `type`, `value` of `delta` properties"
//	@Success		200		{object}	models.Metrics
//	@Failure		400		{string}	string
//	@Security		ApiKeyAuth
//	@Router			/update [post]
func (sh *ServiceHandlers) updateByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerContentType, contentTypeApplicationJSON)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

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
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateByURL godoc
//
//	@Tags			Update
//	@Summary		Update metrics data
//	@Description	Create new or update existing metric data.
//	@ID				updateUpdateByURL
//	@Accept			plain
//	@Produce		plain
//	@Param			name	path		string	true	"Metrics' name"						example("test")
//	@Param			type	path		string	true	"Metrics' type (counter or gauge)"	example("gauge")
//	@Param			value	path		string	true	"Metrics' value (integer or float)"	example("1.1")
//	@Success		200		{string}	string
//	@Failure		400		{string}	string
//	@Security		ApiKeyAuth
//	@Router			/update/{type}/{name}/{value} [post]
func (sh *ServiceHandlers) updateByURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerContentType, contentTypeTextPlain)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

	m, err := models.NewMetricsByStrings(
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

// ValueByJSON godoc
//
//	@Tags			Value
//	@Summary		Return metrics
//	@Description	Return metric value
//	@ID				valueValueByJSON
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.Metrics	true	"A JSON object with `id`, `type` properties"
//	@Success		200		{object}	models.Metrics
//	@Failure		400		{string}	string
//	@Failure		404		{string}	string
//	@Security		ApiKeyAuth
//	@Router			/value [post]
func (sh *ServiceHandlers) valueByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerContentType, contentTypeApplicationJSON)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

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
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ValueByURL godoc
//
//	@Tags			Value
//	@Summary		Return metrics
//	@Description	Return metric value
//	@ID				valueValueByURL
//	@Accept			plain
//	@Produce		plain
//	@Param			name	path		string	true	"Metrics' name"						example("test")
//	@Param			type	path		string	true	"Metrics' type (counter or gauge)"	example("gauge")
//	@Success		200		{string}	string
//	@Failure		400		{string}	string
//	@Failure		404		{string}	string
//	@Security		ApiKeyAuth
//	@Router			/value/{type}/{name} [get]
func (sh *ServiceHandlers) valueByURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerContentType, contentTypeTextPlain)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

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
		_, err := fmt.Fprint(w, *m.Delta)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprint(w, *m.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// All godoc
//
//	@Tags			Value
//	@Summary		Return all metrics
//	@Description	Return all metric value
//	@ID				valueAll
//	@Accept			plain
//	@Produce		html
//	@Success		200	{string}	string
//	@Failure		500	{string}	string
//	@Security		ApiKeyAuth
//	@Router			/ [get]
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
			{{ range $s, $v := . }}
			<tr>
				<td>{{ $s }}</td>
				<td>{{ $v }}</td>
			</tr>
			{{end}}
		</table>
	</body>
	
	</html>`

	w.Header().Add(headerContentType, contetntTypeTextHTML)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

	t := template.New("all tmpl")
	t, err := t.Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	data, err := sh.ms.GetAll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Ping godoc
//
//	@Summary		Ping storage
//	@Description	Check storage work
//	@ID				Ping
//	@Accept			plain
//	@Produce		plain
//	@Success		200	{string}	string
//	@Failure		500	{string}	string
//	@Security		ApiKeyAuth
//	@Router			/ping [get]
func (sh *ServiceHandlers) ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerContentType, contentTypeTextPlain)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

	err := sh.ms.PingDB(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

// Updates godoc
//
//	@Tags			Update
//	@Summary		Multiply metrics update
//	@Description	Create new or update existing metrics data.
//	@ID				updateUpdates
//	@Accept			json
//	@Produce		plain
//	@Param			request	body		[]models.Metrics	true	"A JSON objects with `id`, `type`, `value` of `delta` properties"
//	@Success		200		{string}	string
//	@Failure		400		{string}	string
//	@Failure		500		{string}	string
//	@Security		ApiKeyAuth
//	@Router			/updates [post]
func (sh *ServiceHandlers) updates(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerContentType, contentTypeTextPlain)
	w.Header().Add(headerContentType, contentTypeCharsetUTF8)

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := models.NewMetricsSliceByJSON(buf.Bytes())
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

	m, err := models.NewMetricsByJSON(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return m, err
}

// ServiceRouter return router for run server.
func ServiceRouter(gp *compresses.GzipPool, hasher *hasher.Hasher, sh ServiceHandlers, dm Decrypter, ipChecker IPChecker) chi.Router {
	r := chi.NewRouter()
	r.Use(dm.RequestDecrypt)
	r.Use(hasher.RequestHash)
	r.Use(gp.RequestCompress)
	r.Use(ipChecker.RequsetIPCheck)
	r.Use(logger.RequestLogger)
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
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("swagger/doc.json")))
	})

	return r
}
