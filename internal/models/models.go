// Package models defines structure and methods for working with server model.
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Metrics model
// @Description Metric information
// @Description type may be "gauge" or "counter"
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// NewMetrics returns an empty model with the specified name and type.
func NewMetrics(id, mType string) (*Metrics, error) {
	if err := checkType(mType); err != nil {
		return nil, fmt.Errorf("check metrics type id %s, mType %s: %w", id, mType, err)
	}

	return &Metrics{ID: id, MType: mType}, nil
}

// NewMetricsForGauge returns a gauge model with the specified id and value.
func NewMetricsForGauge(id string, value float64) *Metrics {
	return &Metrics{ID: id, MType: "gauge", Value: &value}
}

// NewMetricsForCounter returns a counter model with the specified id and delta.
func NewMetricsForCounter(id string, delta int64) *Metrics {
	return &Metrics{ID: id, MType: "counter", Delta: &delta}
}

// NewMetricsByStrings returns a model with the specified id, type and value.
func NewMetricsByStrings(id, mType, value string) (*Metrics, error) {
	switch strings.ToLower(mType) {
	case "counter":
		return counterMetricsBySting(id, value)
	case "gauge":
		return gaugeMetricsByStrings(id, value)
	default:
		return nil, fmt.Errorf("unknown metrics type name %s, type %s, value %s", id, mType, value)
	}
}

// NewMetricsByJSON returns a model by JSON.
func NewMetricsByJSON(j []byte) (*Metrics, error) {
	var m Metrics
	err := json.Unmarshal(j, &m)

	if err != nil {
		return nil, fmt.Errorf("unmarshal json %s: %w", string(j), err)
	}

	if err := checkType(m.MType); err != nil {
		return nil, fmt.Errorf("check metrics type id %s, mType %s: %w", m.ID, m.MType, err)
	}

	return &m, nil
}

// NewMetricsSliceByJSON returns a set of JSON models.
func NewMetricsSliceByJSON(j []byte) ([]Metrics, error) {
	var m []Metrics
	err := json.Unmarshal(j, &m)

	if err != nil {
		return nil, fmt.Errorf("unmarshal json %s: %w", string(j), err)
	}

	return m, nil
}

// GetGaugesSliceByMap returns models based on the specified data set.
func GetGaugesSliceByMap(m map[string]float64) []Metrics {
	rM := make([]Metrics, 0, len(m))

	for k, v := range m {
		value := v
		rM = append(rM, Metrics{ID: k, MType: "gauge", Value: &value})
	}

	return rM
}

func checkType(mType string) error {
	switch strings.ToLower(mType) {
	case "counter", "gauge":
		return nil
	default:
		return fmt.Errorf("unkonwn type %s", mType)
	}
}

func counterMetricsBySting(id, delta string) (*Metrics, error) {
	v, err := parseCounter(delta)

	if err != nil {
		return nil, fmt.Errorf("parse counter %s %s: %w", id, delta, err)
	}

	return NewMetricsForCounter(id, v), nil
}

func gaugeMetricsByStrings(id, value string) (*Metrics, error) {
	v, err := parseGauge(value)

	if err != nil {
		return nil, fmt.Errorf("parse gauge %s %s: %w", id, value, err)
	}

	return NewMetricsForGauge(id, v), nil
}

func parseGauge(g string) (float64, error) {
	v, err := strconv.ParseFloat(g, 64)

	return v, err
}

func parseCounter(c string) (int64, error) {
	v, err := strconv.ParseInt(c, 10, 64)

	return v, err
}
