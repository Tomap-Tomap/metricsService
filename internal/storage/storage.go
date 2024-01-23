package storage

import (
	"errors"
	"fmt"
	"maps"
	"sync"

	"github.com/DarkOmap/metricsService/internal/models"
)

type Gauge float64
type Counter int64

type gauges struct {
	sync.RWMutex
	data map[string]Gauge
}

type counters struct {
	sync.RWMutex
	data map[string]Counter
}

type MemStorage struct {
	gauges   gauges
	counters counters
}

func NewMemStorage() *MemStorage {
	ms := MemStorage{}
	ms.counters.data = make(map[string]Counter)
	ms.gauges.data = make(map[string]Gauge)

	return &ms
}

func (ms *MemStorage) UpdateByMetrics(m models.Metrics) (models.Metrics, error) {
	switch m.MType {
	case "counter":
		return ms.updateCounterByMetrics(m.ID, (*Counter)(m.Delta))
	case "gauge":
		return ms.updateGaugeByMetrics(m.ID, (*Gauge)(m.Value))
	default:
		return models.Metrics{}, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (ms *MemStorage) updateCounterByMetrics(id string, delta *Counter) (models.Metrics, error) {
	if delta == nil {
		return models.Metrics{}, fmt.Errorf("delta is empty")
	}

	newDelta := int64(ms.addCounter(*delta, id))

	return models.NewMetricsForCounter(id, newDelta), nil
}

func (ms *MemStorage) updateGaugeByMetrics(id string, value *Gauge) (models.Metrics, error) {
	if value == nil {
		return models.Metrics{}, fmt.Errorf("value is empty")
	}

	newValue := float64(ms.setGauge(*value, id))

	return models.NewMetricsForGauge(id, newValue), nil
}

func (ms *MemStorage) ValueByMetrics(m models.Metrics) (models.Metrics, error) {
	switch m.MType {
	case "counter":
		return ms.valueCounterByMetrics(m.ID)
	case "gauge":
		return ms.valueGaugeByMetrics(m.ID)
	default:
		return m, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (ms *MemStorage) valueCounterByMetrics(id string) (models.Metrics, error) {
	c, err := ms.getCounter(id)

	if err != nil {
		return models.Metrics{}, fmt.Errorf("get counter %s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, int64(c)), nil
}

func (ms *MemStorage) valueGaugeByMetrics(id string) (models.Metrics, error) {
	g, err := ms.getGauge(id)

	if err != nil {
		return models.Metrics{}, fmt.Errorf("get gauge %s: %w", id, err)
	}

	return models.NewMetricsForGauge(id, float64(g)), nil
}

func (ms *MemStorage) setGauge(g Gauge, name string) Gauge {
	ms.gauges.Lock()
	ms.gauges.data[name] = g
	retV := ms.gauges.data[name]
	ms.gauges.Unlock()

	return retV
}

func (ms *MemStorage) getGauge(name string) (Gauge, error) {
	ms.gauges.RLock()
	v, ok := ms.gauges.data[name]
	ms.gauges.RUnlock()

	if !ok {
		return v, errors.New("value not found")
	}

	return v, nil
}

func (ms *MemStorage) addCounter(c Counter, name string) Counter {
	ms.counters.Lock()
	ms.counters.data[name] += c
	retC := ms.counters.data[name]
	ms.counters.Unlock()

	return retC
}

func (ms *MemStorage) getCounter(name string) (Counter, error) {
	ms.counters.RLock()
	v, ok := ms.counters.data[name]
	ms.counters.RUnlock()

	if !ok {
		return v, errors.New("value not found")
	}

	return v, nil
}

func (ms *MemStorage) GetAllGauge() (retMap map[string]Gauge) {
	ms.gauges.RLock()
	retMap = maps.Clone(ms.gauges.data)
	ms.gauges.RUnlock()
	return
}

func (ms *MemStorage) GetAllCounter() (retMap map[string]Counter) {
	ms.counters.RLock()
	retMap = maps.Clone(ms.counters.data)
	ms.counters.RUnlock()
	return
}
