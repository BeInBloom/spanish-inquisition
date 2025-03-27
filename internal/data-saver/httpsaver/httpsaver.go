package httpsaver

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/client-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
)

var (
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrSendingEmptyBatch = errors.New("sending empty batch")
)

type httpSaver struct {
	client    *http.Client
	urlToSend string
	key       string
}

func New(config config.SaverConfig) *httpSaver {
	return &httpSaver{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		//Fix it
		urlToSend: "http://" + config.URL,
		key:       config.Key,
	}
}

// "/update/%s/%s/%s"
// Меня терзают смутные сомнения о том, что код, который имеет альтернативную отправку должен быть "забыт"
// Возможно, стоит сделать возможность выбора или механизм выбора альтернативного отправления
func (s *httpSaver) Save(data ...models.Metrics) error {
	const fn = "httpSaver.Save"

	if len(data) == 1 {
		return s.sendByParams(data[0])
	}

	if len(data) > 1 {
		return s.sendBatch(data)
	}

	return ErrSendingEmptyBatch
}

func (s *httpSaver) sendBatch(data []models.Metrics) error {
	const (
		fn          = "httpSaver.sendBatch"
		batchSuffix = "/updates/"
	)

	jsonMetric, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
	}

	reqAddr := s.urlToSend + batchSuffix
	req, err := http.NewRequest(http.MethodPost, reqAddr, bytes.NewBuffer(jsonMetric))
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	if s.key != "" {
		hash := sha256.Sum256(jsonMetric)
		req.Header.Set("HashSHA256", string(hash[:]))
	}

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

func (s *httpSaver) sendByParams(data models.Metrics) error {
	const (
		fn           = "httpSaver.sendByParams"
		updateSuffix = "/update/"
	)
	reqString, err := s.getStringByModel(data)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	res, err := s.client.Post(fmt.Sprintf(s.urlToSend+updateSuffix+"%s/", reqString), "text/plain", nil)
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

func (s *httpSaver) getStringByModel(data models.Metrics) (string, error) {
	switch data.MType {
	case models.Gauge:
		return fmt.Sprintf("%s/%s/%f", data.MType, data.ID, *data.Value), nil
	case models.Counter:
		return fmt.Sprintf("%s/%s/%d", data.MType, data.ID, *data.Delta), nil
	default:
		return "", ErrInvalidMetricType
	}
}

func (s *httpSaver) sendByJSON(data models.Metrics) error {
	const fn = "httpSaver.sendByJSON"

	jsonMetric, err := json.Marshal(data)
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
