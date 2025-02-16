package httpsaver

import (
	"testing"
)

func TestHttpSaver_Save_Success(t *testing.T) {
}

func TestHttpSaver_Save_HttpError(t *testing.T) {
	// mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	http.Error(w, "internal error", http.StatusInternalServerError)
	// }))
	// defer mockServer.Close()

	// saverConfig := config.SaverConfig{
	// 	Timeout: 2,
	// 	URL:     mockServer.URL + "/%s/%s/%s",
	// }

	// saver := New(saverConfig)

	// data := ptypes.SendData{
	// 	MetricType: "gauge",
	// 	Name:       "test-metric",
	// 	Value:      "42",
	// }

	// err := saver.Save(data)
	// assert.Error(t, err)
}

func TestHttpSaver_Save_Timeout(t *testing.T) {
	// TODO
}
