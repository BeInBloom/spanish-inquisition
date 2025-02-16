package models

import (
	"fmt"
	"strconv"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

var (
	ErrUnexpectedMetricType = fmt.Errorf("unexpected metric type")
)

type Metrics struct {
	ID    string   `json:"id" validate:"required"`
	MType string   `json:"type" validate:"required,oneof=gauge counter"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func ParseMetrics(m Metrics) (string, string, string) {
	switch m.MType {
	case Gauge:
		return m.MType, m.ID, fmt.Sprintf("%v", *m.Value)
	case Counter:
		return m.MType, m.ID, fmt.Sprintf("%v", *m.Delta)
	default:
		return "", "", ""
	}
}

func CreateMetricsByType(mType, id, val string) (Metrics, error) {
	switch mType {
	case Gauge:
		return createGauge(id, val)
	case Counter:
		return createCounter(id, val)
	default:
		return Metrics{}, ErrUnexpectedMetricType
	}
}

func createGauge(id, val string) (Metrics, error) {
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return Metrics{}, err
	}

	return Metrics{
		ID:    id,
		MType: Gauge,
		Value: &num,
	}, nil
}

func createCounter(id, val string) (Metrics, error) {
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return Metrics{}, err
	}

	return Metrics{
		ID:    id,
		MType: Counter,
		Delta: &num,
	}, nil
}
