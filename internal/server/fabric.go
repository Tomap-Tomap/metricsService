package server

import (
	"context"
	"fmt"

	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/DarkOmap/metricsService/internal/storage"
)

type Repository interface {
	UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	GetAllGauge(ctx context.Context) (map[string]storage.Gauge, error)
	GetAllCounter(ctx context.Context) (map[string]storage.Counter, error)
	PingDB(ctx context.Context) error
	Updates(ctx context.Context, metrics []models.Metrics) error
	Close()
}

func NewRepository(ctx context.Context, p parameters.ServerParameters) (Repository, error) {
	if p.DataBaseDSN != "" {
		logger.Log.Info("Create database storage")
		r, err := storage.NewDBStorage(ctx, p)

		if err != nil {
			return nil, fmt.Errorf("create database storage: %w", err)
		}

		return r, nil
	}

	logger.Log.Info("Create in memory storage")
	r, err := storage.NewMemStorage(ctx, p)

	if err != nil {
		return nil, fmt.Errorf("create in memory storage: %w", err)
	}

	return r, nil
}
