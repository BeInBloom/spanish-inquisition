package datafetcher

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	h "github.com/BeInBloom/spanish-inquisition/internal/helpers"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrCantFetchData = errors.New("can't fetch data")
)

type fetcherFunc func() ([]models.Metrics, error)

func (f fetcherFunc) Fetch() ([]models.Metrics, error) {
	return f()
}

type fetcher interface {
	Fetch() ([]models.Metrics, error)
}

type dataFetcher struct {
	ctx          context.Context
	data         []models.Metrics
	timeToUpdate int64
	mutex        sync.RWMutex
	running      int64
	fetchers     []fetcher
}

func New(ctx context.Context, timeToUpdate int64) *dataFetcher {
	fetcher := &dataFetcher{
		ctx:          ctx,
		timeToUpdate: timeToUpdate,
		data:         make([]models.Metrics, 0),
		running:      0,
	}

	fetcher.AddFetcher(fetcherFunc(specificMetrics))
	fetcher.AddFetcher(fetcherFunc(gopsutilMetrics))
	fetcher.AddFetcher(fetcherFunc(runtimeMetrics))

	fetcher.start()

	return fetcher
}

func (d *dataFetcher) Fetch() ([]models.Metrics, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.data != nil {
		returnedData := make([]models.Metrics, len(d.data))
		copy(returnedData, d.data)
		return returnedData, nil
	}

	return nil, ErrCantFetchData
}

func (d *dataFetcher) AddFetcher(fetcher fetcher) {
	d.fetchers = append(d.fetchers, fetcher)
}

func (d *dataFetcher) start() {
	const fn = "dataFetcher.start"

	if atomic.CompareAndSwapInt64(&d.running, 1, 0) {
		return
	}

	data, err := d.fetchAll()
	if err == nil {
		d.data = data
	}

	ticker := time.NewTicker(time.Duration(d.timeToUpdate) * time.Second)

	go func() {
		defer func() {
			ticker.Stop()
			atomic.StoreInt64(&d.running, 0)
		}()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				data, err := d.fetchAll()

				d.mutex.Lock()

				if err != nil {
					fmt.Printf("Error fetching data: %v\n", err)
					d.data = nil
				} else {
					d.data = data
				}

				d.mutex.Unlock()
			}
		}
	}()
}

func (d *dataFetcher) fetchAll() ([]models.Metrics, error) {
	const fn = "dataFetcher.fetchAll"

	var allData []models.Metrics
	var allErrors []error
	cData := make(chan []models.Metrics, len(d.fetchers))
	cErr := make(chan error, len(d.fetchers))

	go func() {
		for _, f := range d.fetchers {
			go func(fetcher fetcher) {
				data, err := fetcher.Fetch()
				if err != nil {
					cErr <- err
					return
				}

				cData <- data
			}(f)
		}

		defer func() {
			close(cData)
			close(cErr)
		}()
	}()

	//Два синхронных чтения из канала. Пока мне проще мыслить из серии локнуть мутексом, добавить в слайс
	//и возвращать его. С каналами пока как-то непонятное.
	for data := range cData {
		allData = append(allData, data...)
	}

	for err := range cErr {
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	if len(allErrors) > 0 {
		return nil, fmt.Errorf("%s: %v", fn, allErrors)
	}

	return allData, nil
}

func specificMetrics() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)

	var step int64 = 1
	metrics = append(metrics, models.Metrics{
		MType: Counter,
		ID:    "PollCount",
		Delta: &step,
	})
	metrics = append(metrics, models.Metrics{
		MType: Gauge,
		ID:    "RandomValue",
		Value: h.GetRandomFloat(0, 10000),
	})

	return metrics, nil
}

func runtimeMetrics() ([]models.Metrics, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := []models.Metrics{
		{MType: Gauge, ID: "Alloc", Value: toFloat64Ptr(m.Alloc)},
		{MType: Gauge, ID: "BuckHashSys", Value: toFloat64Ptr(m.BuckHashSys)},
		{MType: Gauge, ID: "Frees", Value: toFloat64Ptr(m.Frees)},
		{MType: Gauge, ID: "GCCPUFraction", Value: toFloat64Ptr(m.GCCPUFraction)},
		{MType: Gauge, ID: "GCSys", Value: toFloat64Ptr(m.GCSys)},
		{MType: Gauge, ID: "HeapAlloc", Value: toFloat64Ptr(m.HeapAlloc)},
		{MType: Gauge, ID: "HeapIdle", Value: toFloat64Ptr(m.HeapIdle)},
		{MType: Gauge, ID: "HeapInuse", Value: toFloat64Ptr(m.HeapInuse)},
		{MType: Gauge, ID: "HeapObjects", Value: toFloat64Ptr(m.HeapObjects)},
		{MType: Gauge, ID: "HeapReleased", Value: toFloat64Ptr(m.HeapReleased)},
		{MType: Gauge, ID: "HeapSys", Value: toFloat64Ptr(m.HeapSys)},
		{MType: Gauge, ID: "LastGC", Value: toFloat64Ptr(m.LastGC)},
		{MType: Gauge, ID: "Lookups", Value: toFloat64Ptr(m.Lookups)},
		{MType: Gauge, ID: "MCacheInuse", Value: toFloat64Ptr(m.MCacheInuse)},
		{MType: Gauge, ID: "MCacheSys", Value: toFloat64Ptr(m.MCacheSys)},
		{MType: Gauge, ID: "MSpanInuse", Value: toFloat64Ptr(m.MSpanInuse)},
		{MType: Gauge, ID: "MSpanSys", Value: toFloat64Ptr(m.MSpanSys)},
		{MType: Gauge, ID: "Mallocs", Value: toFloat64Ptr(m.Mallocs)},
		{MType: Gauge, ID: "NextGC", Value: toFloat64Ptr(m.NextGC)},
		{MType: Gauge, ID: "NumForcedGC", Value: toFloat64Ptr(m.NumForcedGC)},
		{MType: Gauge, ID: "NumGC", Value: toFloat64Ptr(m.NumGC)},
		{MType: Gauge, ID: "OtherSys", Value: toFloat64Ptr(m.OtherSys)},
		{MType: Gauge, ID: "PauseTotalNs", Value: toFloat64Ptr(m.PauseTotalNs)},
		{MType: Gauge, ID: "StackInuse", Value: toFloat64Ptr(m.StackInuse)},
		{MType: Gauge, ID: "StackSys", Value: toFloat64Ptr(m.StackSys)},
		{MType: Gauge, ID: "Sys", Value: toFloat64Ptr(m.Sys)},
		{MType: Gauge, ID: "TotalAlloc", Value: toFloat64Ptr(m.TotalAlloc)},
	}

	return metrics, nil
}

func gopsutilMetrics() ([]models.Metrics, error) {
	var metrics []models.Metrics

	mem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	metrics = append(metrics, models.Metrics{
		MType: Gauge,
		ID:    "MemoryTotal",
		Value: toFloat64Ptr(mem.Total),
	})

	metrics = append(metrics, models.Metrics{
		MType: Gauge,
		ID:    "MemoryUsed",
		Value: toFloat64Ptr(mem.Free),
	})

	cpuCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	metrics = append(metrics, models.Metrics{
		MType: Gauge,
		ID:    "CPUCount",
		Value: toFloat64Ptr(cpuCount),
	})

	return metrics, nil

}

func toFloat64Ptr(value interface{}) *float64 {
	switch v := value.(type) {
	case int:
		result := float64(v)
		return &result
	case float32:
		result := float64(v)
		return &result
	case float64:
		return &v
	case int64:
		result := float64(v)
		return &result
	case int32:
		result := float64(v)
		return &result
	case uint:
		result := float64(v)
		return &result
	case uint8:
		result := float64(v)
		return &result
	case uint16:
		result := float64(v)
		return &result
	case uint32:
		result := float64(v)
		return &result
	case uint64:
		result := float64(v)
		return &result
	}

	return nil
}
