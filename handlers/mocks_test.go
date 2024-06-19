package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/stretchr/testify/mock"
)

type StorageMockedObject struct {
	mock.Mock
}

func (sm *StorageMockedObject) UpdateByMetrics(_ context.Context, m models.Metrics) (*models.Metrics, error) {
	args := sm.Called(m)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Metrics), args.Error(1)
}

func (sm *StorageMockedObject) ValueByMetrics(_ context.Context, m models.Metrics) (*models.Metrics, error) {
	args := sm.Called(m)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Metrics), args.Error(1)
}

func (sm *StorageMockedObject) GetAll(context.Context) (map[string]fmt.Stringer, error) {
	args := sm.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]fmt.Stringer), args.Error(1)
}

func (sm *StorageMockedObject) PingDB(context.Context) error {
	args := sm.Called()

	return args.Error(0)
}

func (sm *StorageMockedObject) Updates(_ context.Context, metrics []models.Metrics) error {
	args := sm.Called(metrics)

	return args.Error(0)
}

type DecrypterMockedObject struct{}

func (dmo *DecrypterMockedObject) RequestDecrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

type IPCheckerMockedObject struct{}

func (icmo *IPCheckerMockedObject) RequsetIPCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
