package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"sync"
	"time"

	"github.com/DarkOmap/metricsService/internal/file"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"golang.org/x/sync/errgroup"
)

type gauges struct {
	Data map[string]Gauge `json:"data"`
	sync.RWMutex
}

type counters struct {
	Data map[string]Counter `json:"data"`
	sync.RWMutex
}

type MemStorage struct {
	producer        *file.Producer
	fileStoragePath string
	Gauges          gauges   `json:"gauges"`
	Counters        counters `json:"counters"`
	storeInterval   uint
}

func NewMemStorage(ctx context.Context, eg *errgroup.Group, producer *file.Producer, p parameters.ServerParameters) (*MemStorage, error) {
	ms := MemStorage{}
	ms.Counters.Data = make(map[string]Counter)
	ms.Gauges.Data = make(map[string]Gauge)
	ms.fileStoragePath = p.FileStoragePath
	ms.storeInterval = p.StoreInterval
	ms.producer = producer

	if p.Restore {
		consumer, err := file.NewConsumer(p.FileStoragePath)

		if err != nil {
			return nil, fmt.Errorf("initializing new consumer: %w", err)
		}
		defer consumer.Close()

		m := &models.Metrics{}

		for err := consumer.Decoder.Decode(m); err != io.EOF; err = consumer.Decoder.Decode(m) {
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("read from file for storage: %w", err)
			}

			if m.MType == "counter" {
				ms.Counters.Data[m.ID] = Counter(*m.Delta)
				continue
			}
			_, err := ms.UpdateByMetrics(context.Background(), *m)

			if err != nil {
				return nil, fmt.Errorf("read from file for storage: %w", err)
			}
		}
		ms.producer.ClearFile()
	}

	if p.StoreInterval != 0 {
		ms.runDumping(ctx, eg)
	}

	return &ms, nil
}

func (ms *MemStorage) runDumping(ctx context.Context, eg *errgroup.Group) {
	eg.Go(func() error {
		for {
			select {
			case <-time.After(time.Duration(ms.storeInterval) * time.Second):
				if err := ms.dumpStorage(); err != nil {
					return fmt.Errorf("dump by interval: %w", err)
				}
			case <-ctx.Done():
				if err := ms.dumpStorage(); err != nil {
					return fmt.Errorf("dump by stop: %w", err)
				}
				logger.Log.Info("Stop sync")
				return nil
			}
		}
	})
}

func (ms *MemStorage) dumpStorage() error {
	allGauges, _ := ms.GetAllGauge(context.Background())
	for idx, val := range allGauges {
		m := models.NewMetricsForGauge(idx, float64(val))
		err := ms.producer.WriteInFile(m)

		if err != nil {
			return fmt.Errorf("write in file: %w", err)
		}
	}

	allCouters, _ := ms.GetAllCounter(context.Background())
	for idx, val := range allCouters {
		m := models.NewMetricsForCounter(idx, int64(val))
		err := ms.producer.WriteInFile(m)

		if err != nil {
			return fmt.Errorf("write in file: %w", err)
		}
	}
	return nil
}

func (ms *MemStorage) UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case "counter":
		return ms.updateCounterByMetrics(m.ID, (*Counter)(m.Delta))
	case "gauge":
		return ms.updateGaugeByMetrics(m.ID, (*Gauge)(m.Value))
	default:
		return nil, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (ms *MemStorage) updateCounterByMetrics(id string, delta *Counter) (*models.Metrics, error) {
	if delta == nil {
		return nil, fmt.Errorf("delta is empty")
	}

	newDelta := int64(ms.addCounter(*delta, id))

	return models.NewMetricsForCounter(id, newDelta), nil
}

func (ms *MemStorage) updateGaugeByMetrics(id string, value *Gauge) (*models.Metrics, error) {
	if value == nil {
		return nil, fmt.Errorf("value is empty")
	}

	newValue := float64(ms.setGauge(*value, id))

	return models.NewMetricsForGauge(id, newValue), nil
}

func (ms *MemStorage) ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case "counter":
		return ms.valueCounterByMetrics(m.ID)
	case "gauge":
		return ms.valueGaugeByMetrics(m.ID)
	default:
		return nil, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (ms *MemStorage) valueCounterByMetrics(id string) (*models.Metrics, error) {
	c, err := ms.getCounter(id)

	if err != nil {
		return nil, fmt.Errorf("get counter %s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, int64(c)), nil
}

func (ms *MemStorage) valueGaugeByMetrics(id string) (*models.Metrics, error) {
	g, err := ms.getGauge(id)

	if err != nil {
		return nil, fmt.Errorf("get gauge %s: %w", id, err)
	}

	return models.NewMetricsForGauge(id, float64(g)), nil
}

func (ms *MemStorage) setGauge(g Gauge, name string) Gauge {
	ms.Gauges.Lock()
	defer ms.Gauges.Unlock()
	ms.Gauges.Data[name] = g
	retV := ms.Gauges.Data[name]

	if ms.storeInterval == 0 {
		m := models.NewMetricsForGauge(name, float64(g))
		ms.producer.WriteInFile(m)
	}

	return retV
}

func (ms *MemStorage) getGauge(name string) (Gauge, error) {
	ms.Gauges.RLock()
	v, ok := ms.Gauges.Data[name]
	ms.Gauges.RUnlock()

	if !ok {
		return v, errors.New("value not found")
	}

	return v, nil
}

func (ms *MemStorage) addCounter(c Counter, name string) Counter {
	ms.Counters.Lock()
	defer ms.Counters.Unlock()
	ms.Counters.Data[name] += c
	retC := ms.Counters.Data[name]

	if ms.storeInterval == 0 {
		m := models.NewMetricsForCounter(name, int64(retC))
		ms.producer.WriteInFile(m)
	}

	return retC
}

func (ms *MemStorage) getCounter(name string) (Counter, error) {
	ms.Counters.RLock()
	v, ok := ms.Counters.Data[name]
	ms.Counters.RUnlock()

	if !ok {
		return v, errors.New("value not found")
	}

	return v, nil
}

func (ms *MemStorage) GetAllGauge(ctx context.Context) (retMap map[string]Gauge, err error) {
	ms.Gauges.RLock()
	retMap = maps.Clone(ms.Gauges.Data)
	ms.Gauges.RUnlock()
	return
}

func (ms *MemStorage) GetAllCounter(ctx context.Context) (retMap map[string]Counter, err error) {
	ms.Counters.RLock()
	retMap = maps.Clone(ms.Counters.Data)
	ms.Counters.RUnlock()
	return
}

func (ms *MemStorage) PingDB(ctx context.Context) error {
	return fmt.Errorf("for this storage type database is not supported")
}

func (ms *MemStorage) Updates(ctx context.Context, metrics []models.Metrics) error {
	for _, val := range metrics {
		switch val.MType {
		case "gauge":
			_, err := ms.updateGaugeByMetrics(val.ID, (*Gauge)(val.Value))

			if err != nil {
				return err
			}
		case "counter":
			_, err := ms.updateCounterByMetrics(val.ID, (*Counter)(val.Delta))

			if err != nil {
				return err
			}
		}
	}

	return nil
}
