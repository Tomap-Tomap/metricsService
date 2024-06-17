package handlers

import (
	"context"
	"fmt"

	"github.com/DarkOmap/metricsService/internal/models"
)

// Repository it's type for work with storages.
type Repository interface {
	UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	GetAll(ctx context.Context) (map[string]fmt.Stringer, error)
	PingDB(ctx context.Context) error
	Updates(ctx context.Context, metrics []models.Metrics) error
}
