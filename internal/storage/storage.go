package storage

import (
	"errors"
	"strconv"
	"strings"
)

const (
	gaugeType = iota
	counterType
)

type typer interface {
	getType() int
}

type gauge float64

func (g gauge) getType() int {
	return gaugeType
}

type counter int64

func (c counter) getType() int {
	return counterType
}

type dataResult struct {
	Name  string
	Value typer
}

type Repositories interface {
	AddValue(value typer, name string) error
	GetValue(valueType int, name string) (typer, error)
	GetData() []dataResult
}

type MemStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

func NewMemStorage() MemStorage {
	ms := MemStorage{}
	ms.counters = make(map[string]counter)
	ms.gauges = make(map[string]gauge)

	return ms
}

func (m *MemStorage) AddValue(value typer, name string) error {
	switch v := value.(type) {
	case gauge:
		m.gauges[name] = v
	case counter:
		m.counters[name] += v
	default:
		return errors.New("metrics type is unknown")
	}

	return nil
}

func (m *MemStorage) GetValue(valueType int, name string) (typer, error) {
	switch valueType {
	case gaugeType:
		v, ok := m.gauges[name]

		if !ok {
			return v, errors.New("value not found")
		}

		return v, nil
	case counterType:
		v, ok := m.counters[name]

		if !ok {
			return v, errors.New("value not found")
		}

		return v, nil
	default:
		return counter(0), errors.New("unknown type")
	}
}

func (m *MemStorage) GetData() []dataResult {
	res := make([]dataResult, 0)

	for k, v := range m.counters {
		res = append(res, dataResult{k, v})
	}

	for k, v := range m.gauges {
		res = append(res, dataResult{k, v})
	}

	return res
}

func ParseType(t string) (int, error) {
	switch strings.ToLower(t) {
	case "gauge":
		return gaugeType, nil
	case "counter":
		return counterType, nil
	default:
		return -1, errors.New("unknown type")
	}
}

func ParseGauge(g string) (gauge, error) {
	v, err := strconv.ParseFloat(g, 64)

	return gauge(v), err
}

func ParseCounter(c string) (counter, error) {
	v, err := strconv.ParseInt(c, 10, 64)

	return counter(v), err
}
