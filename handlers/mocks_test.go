package handlers

import (
	"context"
	"net/http"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/stretchr/testify/mock"
)

type StorageMockedObject struct {
	mock.Mock
}

func (sm *StorageMockedObject) UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	args := sm.Called(m)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Metrics), args.Error(1)
}

func (sm *StorageMockedObject) ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	args := sm.Called(m)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Metrics), args.Error(1)
}

func (sm *StorageMockedObject) GetAllGauge(ctx context.Context) (map[string]storage.Gauge, error) {
	args := sm.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]storage.Gauge), args.Error(1)
}

func (sm *StorageMockedObject) GetAllCounter(ctx context.Context) (map[string]storage.Counter, error) {
	args := sm.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]storage.Counter), args.Error(1)
}

func (sm *StorageMockedObject) PingDB(ctx context.Context) error {
	args := sm.Called()

	return args.Error(0)
}

func (sm *StorageMockedObject) Updates(ctx context.Context, metrics []models.Metrics) error {
	args := sm.Called(metrics)

	return args.Error(0)
}

type DecrypterMockedObject struct {
}

func (dmo *DecrypterMockedObject) RequestDecrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

type IPCheckerMockedObject struct {
}

func (icmo *IPCheckerMockedObject) RequsetIPCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
