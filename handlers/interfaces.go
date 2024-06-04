package handlers

import (
	"context"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/storage"
)

// Repository it's type for work with storages.
type Repository interface {
	UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	GetAllGauge(ctx context.Context) (map[string]storage.Gauge, error)
	GetAllCounter(ctx context.Context) (map[string]storage.Counter, error)
	PingDB(ctx context.Context) error
	Updates(ctx context.Context, metrics []models.Metrics) error
}
