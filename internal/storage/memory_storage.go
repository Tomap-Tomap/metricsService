package storage

import (
	"context"
	"fmt"
	"io"
	"maps"
	"sync"
	"time"

	"github.com/DarkOmap/metricsService/internal/file"
	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"go.uber.org/zap"
)

type gauges struct {
	Data map[string]Gauge `json:"data"`
	sync.RWMutex
}

type counters struct {
	Data map[string]Counter `json:"data"`
	sync.RWMutex
}

// MemStorage it's in-memory storage repository
type MemStorage struct {
	producer      *file.Producer
	Gauges        gauges   `json:"gauges"`
	Counters      counters `json:"counters"`
	storeInterval uint
}

// NewMemStorage allocates new MemStorage with parameters
func NewMemStorage(ctx context.Context, p parameters.ServerParameters) (*MemStorage, error) {
	ms := MemStorage{}

	if p.FileStoragePath != "" {
		producer, err := file.NewProducer(p.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("create file producer: %w", err)
		}

		ms.producer = producer
	}

	ms.Counters.Data = make(map[string]Counter)
	ms.Gauges.Data = make(map[string]Gauge)
	ms.storeInterval = p.StoreInterval

	if p.Restore && p.FileStoragePath != "" {
		consumer, err := file.NewConsumer(p.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("initializing new consumer: %w", err)
		}

		m := &models.Metrics{}

		for err := consumer.Decoder.Decode(m); err != io.EOF; err = consumer.Decoder.Decode(m) {
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("read from file for storage: %w", err)
			}

			if m.MType == models.TypeCounter {
				ms.Counters.Data[m.ID] = Counter(*m.Delta)
				continue
			}
			_, err := ms.UpdateByMetrics(context.Background(), *m)
			if err != nil {
				return nil, fmt.Errorf("update metrics from file: %w", err)
			}
		}
		err = consumer.Close()
		if err != nil {
			return nil, fmt.Errorf("close consumer: %w", err)
		}

		err = ms.producer.ClearFile()
		if err != nil {
			return nil, fmt.Errorf("clear file of producer: %w", err)
		}
	}

	if p.StoreInterval != 0 {
		ms.runDumping(ctx)
	}

	return &ms, nil
}

// Close closes file producer in MemStorage
func (ms *MemStorage) Close() error {
	return ms.producer.Close()
}

func (ms *MemStorage) runDumping(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(time.Duration(ms.storeInterval) * time.Second):
				if err := ms.dumpStorage(); err != nil {
					logger.Log.Warn("Dump by interval", zap.Error(err))
				}
			case <-ctx.Done():
				if err := ms.dumpStorage(); err != nil {
					logger.Log.Warn("Dump by stop", zap.Error(err))
				}
				logger.Log.Info("Stop sync")
				return
			}
		}
	}()
}

func (ms *MemStorage) dumpStorage() error {
	ms.Gauges.RLock()
	ms.Counters.RLock()
	defer ms.Gauges.RUnlock()
	defer ms.Counters.RUnlock()
	allGauges := maps.Clone(ms.Gauges.Data)
	for idx, val := range allGauges {
		m := models.NewMetricsForGauge(idx, float64(val))
		err := ms.producer.WriteInFile(m)
		if err != nil {
			return fmt.Errorf("write gauges in file: %w", err)
		}
	}

	allCouters := maps.Clone(ms.Counters.Data)
	for idx, val := range allCouters {
		m := models.NewMetricsForCounter(idx, int64(val))
		err := ms.producer.WriteInFile(m)
		if err != nil {
			return fmt.Errorf("write counters in file: %w", err)
		}
	}
	return nil
}

// UpdateByMetrics updates metrics values by model
func (ms *MemStorage) UpdateByMetrics(_ context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case models.TypeCounter:
		return ms.updateCounterByMetrics(m.ID, (*Counter)(m.Delta))
	case models.TypeGauge:
		return ms.updateGaugeByMetrics(m.ID, (*Gauge)(m.Value))
	default:
		return nil, ErrUnknownType
	}
}

func (ms *MemStorage) updateCounterByMetrics(id string, delta *Counter) (*models.Metrics, error) {
	if delta == nil {
		return nil, ErrEmptyDelta
	}

	newDelta, err := ms.addCounter(*delta, id)
	if err != nil {
		return nil, fmt.Errorf("add counter: %w", err)
	}

	return models.NewMetricsForCounter(id, int64(newDelta)), nil
}

func (ms *MemStorage) updateGaugeByMetrics(id string, value *Gauge) (*models.Metrics, error) {
	if value == nil {
		return nil, ErrEmptyValue
	}

	newValue, err := ms.setGauge(*value, id)
	if err != nil {
		return nil, fmt.Errorf("set gauge: %w", err)
	}

	return models.NewMetricsForGauge(id, float64(newValue)), nil
}

// ValueByMetrics returns value of metrics by name and type
func (ms *MemStorage) ValueByMetrics(_ context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case models.TypeCounter:
		return ms.valueCounterByMetrics(m.ID)
	case models.TypeGauge:
		return ms.valueGaugeByMetrics(m.ID)
	default:
		return nil, ErrUnknownType
	}
}

func (ms *MemStorage) valueCounterByMetrics(id string) (*models.Metrics, error) {
	c, err := ms.getCounter(id)
	if err != nil {
		return nil, fmt.Errorf("get counter in mem storage%s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, int64(c)), nil
}

func (ms *MemStorage) valueGaugeByMetrics(id string) (*models.Metrics, error) {
	g, err := ms.getGauge(id)
	if err != nil {
		return nil, fmt.Errorf("get gauge in mem storage %s: %w", id, err)
	}

	return models.NewMetricsForGauge(id, float64(g)), nil
}

func (ms *MemStorage) setGauge(g Gauge, name string) (Gauge, error) {
	ms.Gauges.Lock()
	defer ms.Gauges.Unlock()
	ms.Gauges.Data[name] = g
	retV := ms.Gauges.Data[name]

	if ms.storeInterval == 0 {
		m := models.NewMetricsForGauge(name, float64(g))
		err := ms.producer.WriteInFile(m)
		if err != nil {
			return 0, fmt.Errorf("write gauge in file: %w", err)
		}
	}

	return retV, nil
}

func (ms *MemStorage) getGauge(name string) (Gauge, error) {
	ms.Gauges.RLock()
	v, ok := ms.Gauges.Data[name]
	ms.Gauges.RUnlock()

	if !ok {
		return v, ErrNotFound
	}

	return v, nil
}

func (ms *MemStorage) addCounter(c Counter, name string) (Counter, error) {
	ms.Counters.Lock()
	defer ms.Counters.Unlock()
	ms.Counters.Data[name] += c
	retC := ms.Counters.Data[name]

	if ms.storeInterval == 0 {
		m := models.NewMetricsForCounter(name, int64(retC))
		err := ms.producer.WriteInFile(m)

		return 0, fmt.Errorf("write counter in file: %w", err)
	}

	return retC, nil
}

func (ms *MemStorage) getCounter(name string) (Counter, error) {
	ms.Counters.RLock()
	v, ok := ms.Counters.Data[name]
	ms.Counters.RUnlock()

	if !ok {
		return v, ErrNotFound
	}

	return v, nil
}

// GetAll returns all data of metrics from mem storage
func (ms *MemStorage) GetAll(context.Context) (retMap map[string]fmt.Stringer, err error) {
	ms.Gauges.RLock()
	ms.Counters.RLock()
	defer ms.Gauges.RUnlock()
	defer ms.Counters.RUnlock()
	retMap = make(map[string]fmt.Stringer)
	for k, v := range ms.Gauges.Data {
		retMap[k] = v
	}

	for k, v := range ms.Counters.Data {
		retMap[k] = v
	}

	return
}

// PingDB returns error "for this storage type database is not supported"
func (ms *MemStorage) PingDB(context.Context) error {
	return fmt.Errorf("for this storage type database is not supported")
}

// Updates updates metrics on memstorage
func (ms *MemStorage) Updates(_ context.Context, metrics []models.Metrics) error {
	for _, val := range metrics {
		switch val.MType {
		case models.TypeGauge:
			_, err := ms.updateGaugeByMetrics(val.ID, (*Gauge)(val.Value))
			if err != nil {
				return err
			}
		case models.TypeCounter:
			_, err := ms.updateCounterByMetrics(val.ID, (*Counter)(val.Delta))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
