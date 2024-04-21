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

// GetCompressedJSON return compress JSON data.
func GetCompressedJSON(m any) ([]byte, error) {
	j, err := json.Marshal(m)

	if err != nil {
		return nil, fmt.Errorf("failed model marshal: %w", err)
	}

	var buf bytes.Buffer

	g := gzip.NewWriter(&buf)

	if _, err := g.Write(j); err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	if err = g.Close(); err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return buf.Bytes(), nil
}
