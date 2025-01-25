package ptypes

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type SendData struct {
	MetricType string
	Name       string
	Value      string
}

type Metrics struct {
	Type   string
	Values []Metric
}

type Metric struct {
	Name  string
	Value string
}
