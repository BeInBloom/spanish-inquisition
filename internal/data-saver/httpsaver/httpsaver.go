package httpsaver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/client-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
)

var (
	ErrInvalidMetricType = errors.New("invalid metric type")
)

type httpSaver struct {
	client    *http.Client
	urlToSend string
}

func New(config config.SaverConfig) *httpSaver {
	return &httpSaver{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		//Fix it
		urlToSend: "http://" + config.URL + "/update/",
	}
}

// "/update/%s/%s/%s"
// Меня терзают смутные сомнения о том, что код, который имеет альтернативную отправку должен быть "забыт"
// Возможно, стоит сделать возможность выбора или механизм выбора альтернативного отправления
func (s *httpSaver) Save(data ptypes.SendData) error {
	const fn = "httpSaver.Save"

	if err := s.sendByJSON(data); err != nil {
		if err := s.sendByParams(data); err != nil {
			return fmt.Errorf("%s: %v", fn, err)
		}
	}

	return nil
}

func (s *httpSaver) sendByParams(data ptypes.SendData) error {
	const fn = "httpSaver.sendByParams"

	res, err := s.client.Post(fmt.Sprintf(s.urlToSend+"%s/%s/%s", data.MetricType, data.Name, data.Value), "text/plain", nil)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: unexpected status code: %v", fn, res.StatusCode)
	}

	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()

	return nil
}

func (s *httpSaver) sendByJSON(data ptypes.SendData) error {
	const fn = "httpSaver.sendByJSON"

	m, err := makeMetricModel(data)

	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	jsonMetric, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	req, err := http.NewRequest(http.MethodPost, s.urlToSend, bytes.NewBuffer(jsonMetric))
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: unexpected status code: %v", fn, res.StatusCode)
	}

	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()

	return nil
}

// Мб этот код должен быть вынесен в конверторы, куда-то в модели
// Это пиздец порнография
func makeMetricModel(data ptypes.SendData) (models.Metrics, error) {
	switch data.MetricType {
	case ptypes.Gauge:
		return makeGaugeModel(data)
	case ptypes.Counter:
		return makeCounterModel(data)
	default:
		return models.Metrics{}, ErrInvalidMetricType
	}
}

func makeCounterModel(data ptypes.SendData) (models.Metrics, error) {
	num, err := strconv.ParseInt(data.Value, 10, 64)
	if err != nil {
		return models.Metrics{}, err
	}

	return models.Metrics{
		ID:    data.Name,
		MType: data.MetricType,
		Delta: &num,
	}, nil
}

func makeGaugeModel(data ptypes.SendData) (models.Metrics, error) {
	num, err := strconv.ParseFloat(data.Value, 64)
	if err != nil {
		return models.Metrics{}, err
	}

	return models.Metrics{
		ID:    data.Name,
		MType: data.MetricType,
		Value: &num,
	}, nil
}
