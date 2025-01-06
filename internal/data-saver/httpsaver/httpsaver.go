package httpsaver

import (
	"fmt"
	"net/http"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/client-config"
	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
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
		urlToSend: "http://" + config.URL + "/update/%s/%s/%s",
	}
}

func (s *httpSaver) Save(data ptypes.SendData) error {
	const fn = "httpSaver.Save"

	res, err := s.client.Post(fmt.Sprintf(s.urlToSend, data.MetricType, data.Name, data.Value), "text/plain", nil)
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
