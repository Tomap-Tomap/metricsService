package main

import (
	"context"
	"fmt"
	"net/http/httptest"

	"github.com/DarkOmap/metricsService/internal/handlers"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/mock"
)

type StorageMockedObject struct {
	mock.Mock
}

func (sm *StorageMockedObject) UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	args := sm.Called(m)

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

	return args.Get(0).(map[string]storage.Gauge), args.Error(1)
}

func (sm *StorageMockedObject) GetAllCounter(ctx context.Context) (map[string]storage.Counter, error) {
	args := sm.Called()

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

func Example() {
	s := NewServer()
	defer s.Close()

	r := resty.New().R().SetBody(`{
		"id": "test",
		"type": "gauge",
		"value": 1.1
	}`).SetHeader("Content-Type", "application/json")

	resp, _ := r.Post(s.URL + "/update")

	fmt.Println(resp.StatusCode())
	fmt.Println(resp.Header().Values("Content-Type"))
	fmt.Println(resp.String())

	// Output:
	// 200
	// [application/json charset=utf-8]
	// {"id":"test","type":"gauge","value":1.1}
}

func NewServer() *httptest.Server {
	testGauge := 1.1
	ms := new(StorageMockedObject)

	ms.On("UpdateByMetrics", models.Metrics{
		ID:    "test",
		MType: "gauge",
		Value: &testGauge,
	}).Return(&models.Metrics{
		ID:    "test",
		MType: "gauge",
		Value: &testGauge,
	}, nil)

	sh := handlers.NewServiceHandlers(ms)
	r := handlers.ServiceRouter(sh, "")

	return httptest.NewServer(r)
}
