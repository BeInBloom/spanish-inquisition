package models

import (
	"fmt"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
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
