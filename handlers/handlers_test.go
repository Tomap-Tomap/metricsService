package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkOmap/metricsService/internal/compresses"
	"github.com/DarkOmap/metricsService/internal/hasher"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	textCT = "text/plain; charset=utf-8"
	jsonCT = "application/json; charset=utf-8"
)

func TestServiceHandlers_updateByJSON(t *testing.T) {
	ms := new(StorageMockedObject)
	ms.On("UpdateByMetrics", *models.NewMetricsForGauge("test", 1.1)).Return(models.NewMetricsForGauge("test", 1.1), nil)
	ms.On("UpdateByMetrics", *models.NewMetricsForCounter("test", 1)).Return(models.NewMetricsForCounter("test", 1), nil)

	dmo := new(DecrypterMockedObject)

	ipcmo := new(IPCheckerMockedObject)

	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		contentType, body string
		code              int
	}
	type param struct {
		method, body string
	}
	tests := []struct {
		name  string
		param param
		want  want
	}{
		{
			name:  "method get",
			param: param{method: http.MethodGet},
			want:  want{code: http.StatusMethodNotAllowed, contentType: ""},
		},
		{
			name: "wrong type value",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "type",
				"value": 0
			}`},
			want: want{code: http.StatusBadRequest, contentType: textCT},
		},
		{
			name: "wrong gauge value",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "gauge",
				"value": "0"
			}`},
			want: want{code: http.StatusBadRequest, contentType: textCT},
		},
		{
			name: "wrong counter value",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "counter",
				"delta": "0"
			}`},
			want: want{code: http.StatusBadRequest, contentType: textCT},
		},
		{
			name: "positive gauge",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "gauge",
				"value": 1.1
			}`},
			want: want{code: http.StatusOK, contentType: jsonCT, body: `{
				"id": "test",
				"type": "gauge",
				"value": 1.1
			}`},
		},
		{
			name: "positive counter",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "counter",
				"delta": 1
			}`},
			want: want{code: http.StatusOK, contentType: jsonCT, body: `{
				"id": "test",
				"type": "counter",
				"delta": 1
			}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := testRequest(t, srv, tt.param.method, "/update", tt.param.body)
			assert.Equal(t, tt.want.code, res.StatusCode())
			assert.Equal(t, tt.want.contentType, strings.Join(res.Header().Values("Content-Type"), "; "))
			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, string(res.Body()))
			}
		})
	}

	ms.AssertExpectations(t)

	t.Run("error update by metrics", func(t *testing.T) {
		ms := new(StorageMockedObject)
		ms.On("UpdateByMetrics", *models.NewMetricsForCounter("test", 1)).Return(nil, fmt.Errorf("test error"))

		dmo := new(DecrypterMockedObject)
		ipcmo := new(IPCheckerMockedObject)
		sh := NewServiceHandlers(ms)
		h := hasher.NewHasher(make([]byte, 0), 1)
		r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

		srv := httptest.NewServer(r)
		defer srv.Close()

		res := testRequest(t, srv, http.MethodPost, "/update",
			`{
			"id": "test",
			"type": "counter",
			"delta": 1
			}`,
		)

		require.Equal(t, 400, res.StatusCode())
		require.Equal(t, "test error\n", string(res.Body()))

		ms.AssertExpectations(t)
	})
}

func TestServiceHandlers_updateByURL(t *testing.T) {
	ms := new(StorageMockedObject)
	ms.On("UpdateByMetrics", *models.NewMetricsForGauge("test", 12)).Return(models.NewMetricsForGauge("test", 12), nil)
	ms.On("UpdateByMetrics", *models.NewMetricsForCounter("test", 12)).Return(models.NewMetricsForCounter("test", 12), nil)

	dmo := new(DecrypterMockedObject)
	ipcmo := new(IPCheckerMockedObject)
	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		contentType string
		code        int
	}
	type param struct {
		method, url string
	}
	tests := []struct {
		name  string
		param param
		want  want
	}{
		{
			name:  "method get",
			param: param{http.MethodGet, "/update/gauge/test/123"},
			want:  want{"", http.StatusMethodNotAllowed},
		},
		{
			name:  "short url",
			param: param{http.MethodPost, "/update/gauge/test"},
			want:  want{textCT, http.StatusNotFound},
		},
		{
			name:  "wrong gauge value",
			param: param{http.MethodPost, "/update/gauge/test/wrong"},
			want:  want{textCT, http.StatusBadRequest},
		},
		{
			name:  "wrong counter value",
			param: param{http.MethodPost, "/update/counter/test/wrong"},
			want:  want{textCT, http.StatusBadRequest},
		},
		{
			name:  "positive gauge",
			param: param{http.MethodPost, "/update/gauge/test/12"},
			want:  want{textCT, http.StatusOK},
		},
		{
			name:  "positive counter",
			param: param{http.MethodPost, "/update/counter/test/12"},
			want:  want{textCT, http.StatusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.param.method
			req.URL = srv.URL + tt.param.url

			res := testRequest(t, srv, tt.param.method, tt.param.url, "")
			assert.Equal(t, tt.want.code, res.StatusCode())
			assert.Equal(t, tt.want.contentType, strings.Join(res.Header().Values("Content-Type"), "; "))
		})
	}

	ms.AssertExpectations(t)

	t.Run("error update by metrics", func(t *testing.T) {
		ms := new(StorageMockedObject)
		ms.On("UpdateByMetrics", *models.NewMetricsForCounter("test", 1)).Return(nil, fmt.Errorf("test error"))

		dmo := new(DecrypterMockedObject)
		ipcmo := new(IPCheckerMockedObject)
		sh := NewServiceHandlers(ms)
		h := hasher.NewHasher(make([]byte, 0), 1)
		r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

		srv := httptest.NewServer(r)
		defer srv.Close()

		res := testRequest(t, srv, http.MethodPost, "/update/counter/test/1", "")

		require.Equal(t, 400, res.StatusCode())
		require.Equal(t, "test error\n", string(res.Body()))

		ms.AssertExpectations(t)
	})
}

func testRequest(t *testing.T, srv *httptest.Server, method, url string, body string) *resty.Response {
	req := resty.New().R()
	req.Method = method
	req.URL = srv.URL + url
	req.SetBody(body)
	res, err := req.Send()
	assert.NoError(t, err)

	return res
}

func TestServiceHandlers_valueByURL(t *testing.T) {
	ms := new(StorageMockedObject)

	testWrong, err := models.NewMetrics("wrong", "counter")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testWrong).Return(nil, errors.New("unknown type wrong"))

	testGauge, err := models.NewMetrics("test", "gauge")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testGauge).Return(models.NewMetricsForGauge("test", 1.1), nil)

	testCounter, err := models.NewMetrics("test", "counter")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testCounter).Return(models.NewMetricsForCounter("test", 1), nil)

	dmo := new(DecrypterMockedObject)
	ipcmo := new(IPCheckerMockedObject)
	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		contentType string
		value       string
		code        int
	}
	type param struct {
		method, url string
	}
	tests := []struct {
		name  string
		param param
		want  want
	}{
		{
			name:  "method post",
			param: param{http.MethodPost, "/value/gauge/test"},
			want:  want{code: http.StatusMethodNotAllowed},
		},
		{
			name:  "short url",
			param: param{http.MethodGet, "/value/gauge"},
			want:  want{code: http.StatusNotFound, contentType: textCT},
		},
		{
			name:  "wrong type",
			param: param{http.MethodGet, "/value/wrong/test"},
			want:  want{code: http.StatusNotFound, contentType: textCT},
		},
		{
			name:  "wrong name",
			param: param{http.MethodGet, "/value/counter/wrong"},
			want:  want{code: http.StatusNotFound, contentType: textCT},
		},
		{
			name:  "positive gauge",
			param: param{http.MethodGet, "/value/gauge/test"},
			want:  want{textCT, "1.1", http.StatusOK},
		},
		{
			name:  "positive counter",
			param: param{http.MethodGet, "/value/counter/test"},
			want:  want{textCT, "1", http.StatusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.param.method
			req.URL = srv.URL + tt.param.url

			res := testRequest(t, srv, tt.param.method, tt.param.url, "")
			assert.Equal(t, tt.want.code, res.StatusCode())

			if res.StatusCode() != http.StatusOK {
				return
			}

			assert.Equal(t, tt.want.contentType, strings.Join(res.Header().Values("Content-Type"), "; "))
			assert.Equal(t, tt.want.value, res.String())
		})
	}

	ms.AssertExpectations(t)
}

func TestServiceHandlers_all(t *testing.T) {
	ms := new(StorageMockedObject)
	ms.On("GetAll").Return(map[string]fmt.Stringer{
		"testG": storage.Gauge(1.1),
		"testC": storage.Counter(1),
	}, nil)

	dmo := new(DecrypterMockedObject)
	ipcmo := new(IPCheckerMockedObject)
	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	htmlText := `<!DOCTYPE html>
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
			<tr>
				<td>testC</td>
				<td>1</td>
			</tr>
			<tr>
				<td>testG</td>
				<td>1.100000</td>
			</tr>
		</table>
	</body>
	
	</html>`

	type want struct {
		contentType string
		value       string
		code        int
	}
	type param struct {
		method, url string
	}
	tests := []struct {
		name  string
		param param
		want  want
	}{
		{
			name:  "method post",
			param: param{http.MethodPost, "/"},
			want:  want{code: http.StatusMethodNotAllowed},
		},
		{
			name:  "positive",
			param: param{http.MethodGet, "/"},
			want:  want{"text/html; charset=utf-8", htmlText, http.StatusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.param.method
			req.URL = srv.URL + tt.param.url

			res := testRequest(t, srv, tt.param.method, tt.param.url, "")
			assert.Equal(t, tt.want.code, res.StatusCode())

			if res.StatusCode() != http.StatusOK {
				return
			}

			assert.Equal(t, tt.want.contentType, strings.Join(res.Header().Values("Content-Type"), "; "))
			assert.Equal(t, strings.ReplaceAll(strings.ReplaceAll(tt.want.value, "\t", ""), "\n", ""), strings.ReplaceAll(strings.ReplaceAll(res.String(), "\t", ""), "\n", ""))
		})
	}

	ms.AssertExpectations(t)

	t.Run("error get all counter", func(t *testing.T) {
		ms := new(StorageMockedObject)
		ms.On("GetAll").Return(nil, fmt.Errorf("test error"))

		dmo := new(DecrypterMockedObject)
		ipcmo := new(IPCheckerMockedObject)
		sh := NewServiceHandlers(ms)
		h := hasher.NewHasher(make([]byte, 0), 1)
		r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

		srv := httptest.NewServer(r)
		defer srv.Close()

		res := testRequest(t, srv, http.MethodGet, "/", "")

		require.Equal(t, 500, res.StatusCode())
		require.Equal(t, "test error\n", string(res.Body()))

		ms.AssertExpectations(t)
	})
}

func TestServiceHandlers_valueByJSON(t *testing.T) {
	ms := new(StorageMockedObject)

	testCounterNotFound, err := models.NewMetrics("notFound", "counter")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testCounterNotFound).Return(nil, errors.New("error"))

	testGaugeNotFound, err := models.NewMetrics("notFound", "gauge")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testGaugeNotFound).Return(nil, errors.New("error"))

	testGauge, err := models.NewMetrics("test", "gauge")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testGauge).Return(models.NewMetricsForGauge("test", 1.1), nil)

	testCounter, err := models.NewMetrics("test", "counter")
	require.NoError(t, err)
	ms.On("ValueByMetrics", *testCounter).Return(models.NewMetricsForCounter("test", 1), nil)

	dmo := new(DecrypterMockedObject)
	ipcmo := new(IPCheckerMockedObject)
	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		contentType, body string
		code              int
	}
	type param struct {
		method, body string
	}
	tests := []struct {
		name  string
		param param
		want  want
	}{
		{
			name:  "method get",
			param: param{method: http.MethodGet},
			want:  want{code: http.StatusMethodNotAllowed, contentType: ""},
		},
		{
			name: "gauge not found",
			param: param{method: http.MethodPost, body: `{
				"id": "notFound",
				"type": "gauge"
			}`},
			want: want{code: http.StatusNotFound, contentType: textCT},
		},
		{
			name: "counter not found",
			param: param{method: http.MethodPost, body: `{
				"id": "notFound",
				"type": "counter"
			}`},
			want: want{code: http.StatusNotFound, contentType: textCT},
		},
		{
			name: "positive gauge",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "gauge"
			}`},
			want: want{code: http.StatusOK, contentType: jsonCT, body: `{
				"id": "test",
				"type": "gauge",
				"value": 1.1
			}`},
		},
		{
			name: "positive counter",
			param: param{method: http.MethodPost, body: `{
				"id": "test",
				"type": "counter"
			}`},
			want: want{code: http.StatusOK, contentType: jsonCT, body: `{
				"id": "test",
				"type": "counter",
				"delta": 1
			}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := testRequest(t, srv, tt.param.method, "/value", tt.param.body)
			assert.Equal(t, tt.want.code, res.StatusCode())
			assert.Equal(t, tt.want.contentType, strings.Join(res.Header().Values("Content-Type"), "; "))
			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, string(res.Body()))
			}
		})
	}

	ms.AssertExpectations(t)

	t.Run("error body", func(t *testing.T) {
		ms := new(StorageMockedObject)
		dmo := new(DecrypterMockedObject)

		sh := NewServiceHandlers(ms)
		ipcmo := new(IPCheckerMockedObject)
		h := hasher.NewHasher(make([]byte, 0), 1)
		r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

		srv := httptest.NewServer(r)
		defer srv.Close()

		res := testRequest(t, srv, http.MethodPost, "/value", "error")
		require.Equal(t, 400, res.StatusCode())
	})
}

func TestServiceHandlers_ping(t *testing.T) {
	ms := new(StorageMockedObject)
	dmo := new(DecrypterMockedObject)
	ipcmo := new(IPCheckerMockedObject)
	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	ms.On("PingDB").Return(nil)

	t.Run("test positive", func(t *testing.T) {
		res := testRequest(t, srv, http.MethodGet, "/ping", "")
		assert.Equal(t, http.StatusOK, res.StatusCode())
	})
	ms.AssertExpectations(t)
	ms.On("PingDB").Unset()
	ms.On("PingDB").Return(errors.New("test"))

	t.Run("test negative", func(t *testing.T) {
		res := testRequest(t, srv, http.MethodGet, "/ping", "")
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
	})
	ms.AssertExpectations(t)
}

func TestServiceHandlers_updates(t *testing.T) {
	ms := new(StorageMockedObject)
	dmo := new(DecrypterMockedObject)
	ipcmo := new(IPCheckerMockedObject)
	sh := NewServiceHandlers(ms)
	h := hasher.NewHasher(make([]byte, 0), 1)
	r := ServiceRouter(compresses.NewGzipPool(1), h, sh, dmo, ipcmo)

	srv := httptest.NewServer(r)
	defer srv.Close()

	t.Run("bad body", func(t *testing.T) {
		res := testRequest(t, srv, http.MethodPost, "/updates", "")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode())
	})

	badModelJSON := `
	[
		{
			"id": "error",
			"type": "counter",
			"delta": 1
		}
	]
	`
	badModel, err := models.NewMetricsSliceByJSON([]byte(badModelJSON))
	require.NoError(t, err)

	ms.On("Updates", badModel).Return(fmt.Errorf("test error"))

	goodModelJSON := `
	[
		{
			"id": "error",
			"type": "counter",
			"delta": 2
		}
	]
	`

	goodModel, err := models.NewMetricsSliceByJSON([]byte(goodModelJSON))
	require.NoError(t, err)

	ms.On("Updates", goodModel).Return(nil)

	t.Run("test 500", func(t *testing.T) {
		res := testRequest(t, srv, http.MethodPost, "/updates", badModelJSON)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
	})

	t.Run("test 200", func(t *testing.T) {
		res := testRequest(t, srv, http.MethodPost, "/updates", goodModelJSON)
		assert.Equal(t, http.StatusOK, res.StatusCode())
	})
}
