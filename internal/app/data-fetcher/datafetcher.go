package datafetcher

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	h "github.com/BeInBloom/spanish-inquisition/internal/helpers"
	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrCantFetchData = errors.New("can't fetch data")
)

type dataFetcher struct {
	ctx          context.Context
	data         []ptypes.SendData
	timeToUpdate int64
	mutex        sync.RWMutex
	running      int64
}

func New(ctx context.Context, timeToUpdate int64) *dataFetcher {
	fetcher := &dataFetcher{
		ctx:          ctx,
		timeToUpdate: timeToUpdate,
		data:         make([]ptypes.SendData, 0),
		running:      0,
	}

	fetcher.start()

	return fetcher
}

func (d *dataFetcher) Fetch() ([]ptypes.SendData, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.data != nil {
		returnedData := make([]ptypes.SendData, len(d.data))
		copy(returnedData, d.data)
		return returnedData, nil
	}

	return nil, ErrCantFetchData
}

func (d *dataFetcher) start() {
	if atomic.CompareAndSwapInt64(&d.running, 1, 0) {
		return
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

func (d *dataFetcher) fetchAll() ([]ptypes.SendData, error) {
	metrics := d.fetchMetrics()
	specificMetrics := d.fetchSpecificMetrics()

	return append(metrics, specificMetrics...), nil
}

func (d *dataFetcher) fetchSpecificMetrics() []ptypes.SendData {
	metrics := make([]ptypes.SendData, 0)

	metrics = append(metrics, ptypes.SendData{
		MetricType: Counter,
		Name:       "PollCount",
		Value:      "1",
	})
	metrics = append(metrics, ptypes.SendData{
		MetricType: Gauge,
		Name:       "RandomValue",
		Value:      fmt.Sprintf("%v", h.GetRandomFloat(0, 10000)),
	})

	return metrics
}

func (d *dataFetcher) fetchMetrics() []ptypes.SendData {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := []ptypes.SendData{
		{MetricType: Gauge, Name: "Alloc", Value: strconv.FormatUint(m.Alloc, 10)},
		{MetricType: Gauge, Name: "BuckHashSys", Value: strconv.FormatUint(m.BuckHashSys, 10)},
		{MetricType: Gauge, Name: "Frees", Value: strconv.FormatUint(m.Frees, 10)},
		{MetricType: Gauge, Name: "GCCPUFraction", Value: strconv.FormatFloat(m.GCCPUFraction, 'f', -1, 64)},
		{MetricType: Gauge, Name: "GCSys", Value: strconv.FormatUint(m.GCSys, 10)},
		{MetricType: Gauge, Name: "HeapAlloc", Value: strconv.FormatUint(m.HeapAlloc, 10)},
		{MetricType: Gauge, Name: "HeapIdle", Value: strconv.FormatUint(m.HeapIdle, 10)},
		{MetricType: Gauge, Name: "HeapInuse", Value: strconv.FormatUint(m.HeapInuse, 10)},
		{MetricType: Gauge, Name: "HeapObjects", Value: strconv.FormatUint(m.HeapObjects, 10)},
		{MetricType: Gauge, Name: "HeapReleased", Value: strconv.FormatUint(m.HeapReleased, 10)},
		{MetricType: Gauge, Name: "HeapSys", Value: strconv.FormatUint(m.HeapSys, 10)},
		{MetricType: Gauge, Name: "LastGC", Value: strconv.FormatUint(m.LastGC, 10)},
		{MetricType: Gauge, Name: "Lookups", Value: strconv.FormatUint(m.Lookups, 10)},
		{MetricType: Gauge, Name: "MCacheInuse", Value: strconv.FormatUint(m.MCacheInuse, 10)},
		{MetricType: Gauge, Name: "MCacheSys", Value: strconv.FormatUint(m.MCacheSys, 10)},
		{MetricType: Gauge, Name: "MSpanInuse", Value: strconv.FormatUint(m.MSpanInuse, 10)},
		{MetricType: Gauge, Name: "MSpanSys", Value: strconv.FormatUint(m.MSpanSys, 10)},
		{MetricType: Gauge, Name: "Mallocs", Value: strconv.FormatUint(m.Mallocs, 10)},
		{MetricType: Gauge, Name: "NextGC", Value: strconv.FormatUint(m.NextGC, 10)},
		{MetricType: Gauge, Name: "NumForcedGC", Value: strconv.FormatUint(uint64(m.NumForcedGC), 10)},
		{MetricType: Gauge, Name: "NumGC", Value: strconv.FormatUint(uint64(m.NumGC), 10)},
		{MetricType: Gauge, Name: "OtherSys", Value: strconv.FormatUint(m.OtherSys, 10)},
		{MetricType: Gauge, Name: "PauseTotalNs", Value: strconv.FormatUint(m.PauseTotalNs, 10)},
		{MetricType: Gauge, Name: "StackInuse", Value: strconv.FormatUint(m.StackInuse, 10)},
		{MetricType: Gauge, Name: "StackSys", Value: strconv.FormatUint(m.StackSys, 10)},
		{MetricType: Gauge, Name: "Sys", Value: strconv.FormatUint(m.Sys, 10)},
		{MetricType: Gauge, Name: "TotalAlloc", Value: strconv.FormatUint(m.TotalAlloc, 10)},
	}

	return metrics
}
