package storage

import (
	"errors"
	"strconv"
	"strings"
)

type Repositories interface {
	addUnit(u StorageUnit) error
}

type gauge float64
type counter int64

type memStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

type StorageUnit struct {
	unitType, name, value string
}

func (m memStorage) AddUnit(u StorageUnit) error {
	switch strings.ToLower(u.unitType) {
	case "gauge":
		g, err := strconv.ParseFloat(u.value, 64)

		if err != nil {
			return err
		}

		m.gauges[u.name] = gauge(g)
	case "counter":
		c, err := strconv.ParseInt(u.value, 10, 64)

		if err != nil {
			return err
		}

		m.counters[u.name] += counter(c)
	default:
		return errors.New("metrics type is unknown")
	}

	return nil
}

func NewStorageUnit(url string) (StorageUnit, error) {
	param := strings.Split(url, "/")

	if len(param) < 3 {
		return StorageUnit{}, errors.New("URI path is to short")
	}

	return StorageUnit{
		unitType: param[0],
		name:     param[1],
		value:    param[2],
	}, nil
}

var Storage memStorage

func init() {
	Storage.counters = make(map[string]counter)
	Storage.gauges = make(map[string]gauge)
}
