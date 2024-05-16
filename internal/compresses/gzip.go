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
	n, err := c.Writer.Write(p)
	return n, err
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
	writerPool chan *gzip.Writer
	readerPool chan *gzip.Reader
}

func NewGzipPool(rateLimit uint) *GzipPool {
	wp := make(chan *gzip.Writer, rateLimit)
	rp := make(chan *gzip.Reader, rateLimit)

	return &GzipPool{wp, rp}
}

func (gp *GzipPool) Close() {
	close(gp.writerPool)
	close(gp.readerPool)
}

// GetCompressedJSON returns compressed JSON data.
func (gp *GzipPool) GetCompressedJSON(m any) ([]byte, error) {
	w, err := gp.getWriter()

	if err != nil {
		return nil, fmt.Errorf("get writer: %w", err)
	}

	defer gp.putWriter(w)

	j, err := json.Marshal(m)

	if err != nil {
		return nil, fmt.Errorf("failed model marshal: %w", err)
	}

	var buf bytes.Buffer

	w.Reset(&buf)

	if _, err := w.Write(j); err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return buf.Bytes(), nil
}

// CompressHandle return handler for middleware.
// Handle may compress and decompress data.
func (gp *GzipPool) CompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			zr, err := gp.getReader()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = zr.Reset(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = &compressReader{ReadCloser: r.Body, Reader: zr}
			defer r.Body.Close()
			defer gp.putReader(zr)
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		switch {
		case supportsGzip:
			w.Header().Set("Content-Encoding", "gzip")
			gw, err := gp.getWriter()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			gw.Reset(w)
			defer gp.putWriter(gw)
			defer gw.Close()

			next.ServeHTTP(&compressWriter{ResponseWriter: w, Writer: gw}, r)
		default:
			next.ServeHTTP(w, r)
		}

	})
}

func (gp *GzipPool) getWriter() (*gzip.Writer, error) {
	select {
	case w, ok := <-gp.writerPool:
		if !ok {
			return nil, fmt.Errorf("pool is closed")
		}

		return w, nil
	default:
	}

	return gzip.NewWriter(nil), nil
}

func (gp *GzipPool) putWriter(w *gzip.Writer) {
	w.Reset(nil)
	select {
	case gp.writerPool <- w:
	default:
	}
}

func (gp *GzipPool) getReader() (*gzip.Reader, error) {
	select {
	case r, ok := <-gp.readerPool:
		if !ok {
			return nil, fmt.Errorf("pool is closed")
		}

		return r, nil
	default:
	}

	return new(gzip.Reader), nil
}

func (gp *GzipPool) putReader(r *gzip.Reader) {
	select {
	case gp.readerPool <- r:
	default:
	}
}
