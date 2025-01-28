package ptypes

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type SendData struct {
	MetricType string `json:"type"`
	Name       string `json:"name"`
	Value      string `json:"value"`
}

type Metrics struct {
	Type   string   `json:"type"`
	Values []Metric `json:"values"`
}

type Metric struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
