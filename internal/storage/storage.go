package storage

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"sync"
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

func (ms *MemStorage) SetGauge(value string, name string) error {
	g, err := parseGauge(value)

	if err != nil {
		return fmt.Errorf("set gauge %s: %w", value, err)
	}

	ms.gauges.Lock()
	ms.gauges.data[name] = g
	ms.gauges.Unlock()

	return nil
}

func (ms *MemStorage) GetGauge(name string) (Gauge, error) {
	ms.gauges.RLock()
	v, ok := ms.gauges.data[name]
	ms.gauges.RUnlock()

	if !ok {
		return v, errors.New("value not found")
	}

	return v, nil
}

func (ms *MemStorage) AddCounter(value string, name string) error {
	c, err := parseCounter(value)

	if err != nil {
		return fmt.Errorf("add counter %s: %w", value, err)
	}

	ms.counters.Lock()
	ms.counters.data[name] += c
	ms.counters.Unlock()

	return nil
}

func (ms *MemStorage) GetCounter(name string) (Counter, error) {
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

func parseGauge(g string) (Gauge, error) {
	v, err := strconv.ParseFloat(g, 64)

	return Gauge(v), err
}

func parseCounter(c string) (Counter, error) {
	v, err := strconv.ParseInt(c, 10, 64)

	return Counter(v), err
}
