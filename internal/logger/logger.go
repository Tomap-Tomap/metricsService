package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	loggingResponseWriter struct {
		http.ResponseWriter
		wroteHeader bool
		code        int
		bytes       int
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	size, err := r.ResponseWriter.Write(b)
	r.bytes += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.code = statusCode
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

var Log *zap.Logger = zap.NewNop()

func Initialize(level string, outputPath string) error {
	lvl, err := zap.ParseAtomicLevel(level)

	if err != nil {
		return fmt.Errorf("parse level %s: %w", level, err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.OutputPaths = []string{outputPath}
	zl, err := cfg.Build()

	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}

	Log = zl

	return nil
}

func RequestLogger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
		}

		defer func() {
			duration := time.Since(start)

			Log.Info("got incoming HTTP request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.String("duration", duration.String()),
				zap.Int("status", lw.code),
				zap.Int("size", lw.bytes),
			)
		}()

		h.ServeHTTP(&lw, r)
	}

	return http.HandlerFunc(logFn)
}

func LogBadRequest(handlerName, uri string, err error) {
	Log.Info("Got incorrect request",
		zap.String("handler", handlerName),
		zap.String("uri", uri),
		zap.Error(err),
	)
}

func LogNotFound(handlerName, uri string, err error) {
	Log.Info("Value not found",
		zap.String("handler", handlerName),
		zap.String("uri", uri),
		zap.Error(err),
	)
}
