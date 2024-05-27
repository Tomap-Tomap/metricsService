package compresses

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestGzipPool_GetCompressedJSON(t *testing.T) {
	t.Run("check compress", func(t *testing.T) {
		checkJSON := `{"test": "test"}`
		pool := NewGzipPool(1)
		defer pool.Close()
		compressJSON, err := pool.GetCompressedJSON(checkJSON)
		require.NoError(t, err)
		require.NotEqual(t, checkJSON, compressJSON)

		compressJSON, err = pool.GetCompressedJSON(checkJSON)
		require.NoError(t, err)
		require.NotEqual(t, checkJSON, compressJSON)
	})

	t.Run("check closed pool", func(t *testing.T) {
		checkJSON := `{"test": "test"}`
		pool := NewGzipPool(1)
		pool.Close()
		_, err := pool.GetCompressedJSON(checkJSON)
		require.Error(t, err)
	})

	t.Run("marshall error", func(t *testing.T) {
		pool := NewGzipPool(1)
		defer pool.Close()

		type Dummy struct {
			Name string
			Next *Dummy
		}
		dummy := Dummy{Name: "Dummy"}
		dummy.Next = &dummy

		_, err := pool.GetCompressedJSON(dummy)
		require.Error(t, err)
	})
}

func BenchmarkGzipPool_GetCompressedJSON(b *testing.B) {
	pool := NewGzipPool(1)
	defer pool.Close()

	testMap := make(map[string]float64)

	for i := 0.1; i < 1000; i++ {
		name := fmt.Sprintf("test%f", i)
		testMap[name] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.GetCompressedJSON(testMap)
	}
}

func BenchmarkCompressHandle(b *testing.B) {
	successBody := `{
        "response": {
            "text": "Извините, я пока ничего не умею"
        },
        "version": "1.0"
    }`

	webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(successBody))
	})
	pool := NewGzipPool(10)
	defer pool.Close()
	handler := pool.RequestCompress(webhook)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	c := resty.New()

	testMap := make(map[string]float64)

	for i := 0.1; i < 1000; i++ {
		name := fmt.Sprintf("test%f", i)
		testMap[name] = i
	}

	requestBody, _ := pool.GetCompressedJSON(testMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.R().SetBody(requestBody).
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			Post(srv.URL)
	}
}

func TestGzipPool_CompressHandle(t *testing.T) {
	requestBody := `{
        "request": {
            "type": "SimpleUtterance",
            "command": "sudo do something"
        },
        "version": "1.0"
    }`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
        "response": {
            "text": "Извините, я пока ничего не умею"
        },
        "version": "1.0"
    }`

	webhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b []byte
		r.Body.Read(b)
		w.Write([]byte(successBody))
	})

	pool := NewGzipPool(1)
	defer pool.Close()
	handler := pool.RequestCompress(webhook)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	c := resty.New()
	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		resp, err := c.R().SetBody(buf).
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.JSONEq(t, successBody, resp.String())
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)

		resp, err := c.R().SetBody(buf).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.JSONEq(t, successBody, resp.String())
	})

	t.Run("check accepts_gzip closed pool", func(t *testing.T) {
		pool := NewGzipPool(1)
		pool.Close()
		handler := pool.RequestCompress(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		buf := bytes.NewBufferString(requestBody)
		c := resty.New()
		resp, err := c.R().SetBody(buf).
			SetHeader("Content-Type", "application/json").
			Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})

	t.Run("check sends_gzip closed pool", func(t *testing.T) {
		pool := NewGzipPool(1)
		pool.Close()
		handler := pool.RequestCompress(webhook)

		srv := httptest.NewServer(handler)
		defer srv.Close()

		c := resty.New()
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		resp, err := c.R().SetBody(buf).
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			Post(srv.URL)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})
}
