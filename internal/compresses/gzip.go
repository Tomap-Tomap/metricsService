// Package compresses defines structures and handles for working with compressed data.
package compresses

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.Writer.Write(p)
}

type compressReader struct {
	io.ReadCloser
	Reader io.ReadCloser
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.Reader.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.ReadCloser.Close(); err != nil {
		return err
	}

	return c.Reader.Close()
}

// CompressHandle return handler for middleware.
// Handle may compress and decompress data.
func CompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			zr, err := gzip.NewReader(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = &compressReader{ReadCloser: r.Body, Reader: zr}
			defer r.Body.Close()
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		switch {
		case supportsGzip:
			w.Header().Set("Content-Encoding", "gzip")
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			next.ServeHTTP(&compressWriter{ResponseWriter: w, Writer: gz}, r)
		default:
			next.ServeHTTP(w, r)
		}

	})
}

type GzipPool struct {
	pool chan *gzip.Writer
}

func NewGzipPool(rateLimit uint) *GzipPool {
	pool := make(chan *gzip.Writer, rateLimit)

	for p := 1; p <= cap(pool); p++ {
		pool <- gzip.NewWriter(nil)
	}

	return &GzipPool{pool}
}

func (gp *GzipPool) getWriter() (*gzip.Writer, error) {
	w, ok := <-gp.pool

	if !ok {
		return nil, fmt.Errorf("pool is closed")
	}

	return w, nil
}

func (gp *GzipPool) putWriter(w *gzip.Writer) {
	w.Reset(nil)
	select {
	case gp.pool <- w:
	default:
	}
}

// GetCompressedJSON return compress JSON data.
func (gp *GzipPool) GetCompressedJSON(m any) ([]byte, error) {
	w, err := gp.getWriter()

	if err != nil {
		return nil, fmt.Errorf("get writer: %w", err)
	}

	j, err := json.Marshal(m)

	if err != nil {
		return nil, fmt.Errorf("failed model marshal: %w", err)
	}

	var buf bytes.Buffer

	w.Reset(&buf)

	if _, err := w.Write(j); err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	gp.putWriter(w)

	return buf.Bytes(), nil
}
