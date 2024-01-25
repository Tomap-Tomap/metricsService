package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DarkOmap/metricsService/internal/file"
	"github.com/DarkOmap/metricsService/internal/models"
)

type Gauge float64
type Counter int64

type gauges struct {
	sync.RWMutex
	Data map[string]Gauge `json:"data"`
}

type counters struct {
	sync.RWMutex
	Data map[string]Counter `json:"data"`
}

type MemStorage struct {
	Gauges    gauges   `json:"gauges"`
	Counters  counters `json:"counters"`
	fileName  string
	storeChan chan struct{}
	storeFunc func() <-chan struct{}
}

func NewMemStorage(storeInterval uint, fileName string) (*MemStorage, error) {
	ms := MemStorage{}
	ms.Counters.Data = make(map[string]Counter)
	ms.Gauges.Data = make(map[string]Gauge)
	ms.fileName = fileName

	ms.initStoreFunc(storeInterval)

	err := ms.runSyncFromFile()

	if err != nil {
		return nil, err
	}

	return &ms, nil
}

func NewMemStorageFromGile(storeInterval uint, fileName string) (*MemStorage, error) {
	consumer, err := file.NewConsumer(fileName)

	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	ms := &MemStorage{}
	ms.Counters.Data = make(map[string]Counter)
	ms.Gauges.Data = make(map[string]Gauge)

	if err := consumer.Decoder.Decode(&ms); err != nil && err != io.EOF {
		return nil, err
	}

	ms.fileName = fileName
	ms.initStoreFunc(storeInterval)

	err = ms.runSyncFromFile()

	if err != nil {
		return nil, err
	}

	return ms, nil
}

func (ms *MemStorage) initStoreFunc(storeInterval uint) {
	if storeInterval == 0 {
		ms.storeChan = make(chan struct{})
		ms.storeFunc = func() <-chan struct{} {
			return ms.storeChan
		}
	} else {
		ms.storeFunc = func() <-chan struct{} {
			<-time.After(time.Duration(storeInterval) * time.Second)
			return make(<-chan struct{})
		}
	}
}

func (ms *MemStorage) runSyncFromFile() error {
	ch := make(chan error)
	go func() {
		producer, err := file.NewProducer(ms.fileName)

		ch <- err

		defer producer.Close()
		loop := true
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		for loop {
			select {
			case <-ms.storeFunc():
				producer.Seek()
				producer.Encoder.Encode(ms)
			case <-ctx.Done():
				producer.Seek()
				producer.Encoder.Encode(ms)
				loop = false
			}
		}
	}()

	if err := <-ch; err != nil {
		return err
	}

	return nil
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
	ms.Gauges.Lock()
	ms.Gauges.Data[name] = g
	retV := ms.Gauges.Data[name]

	if ms.storeChan != nil {
		ms.storeChan <- struct{}{}
	}
	ms.Gauges.Unlock()

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
	ms.Counters.Data[name] += c
	retC := ms.Counters.Data[name]

	if ms.storeChan != nil {
		ms.storeChan <- struct{}{}
	}

	ms.Counters.Unlock()

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

func (ms *MemStorage) GetAllGauge() (retMap map[string]Gauge) {
	ms.Gauges.RLock()
	retMap = maps.Clone(ms.Gauges.Data)
	ms.Gauges.RUnlock()
	return
}

func (ms *MemStorage) GetAllCounter() (retMap map[string]Counter) {
	ms.Counters.RLock()
	retMap = maps.Clone(ms.Counters.Data)
	ms.Counters.RUnlock()
	return
}
