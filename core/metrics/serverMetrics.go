package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type ServerMetrics struct {
	Registry           *prometheus.Registry
	TotalRequestNumber prometheus.Gauge
	RequestDurations   prometheus.Histogram
	RequestNum         int64
}

func NewServerMetrics() *ServerMetrics {
	serverMetrics := &ServerMetrics{}

	serverMetrics.Registry = prometheus.NewRegistry()

	serverMetrics.Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	serverMetrics.Registry.MustRegister(collectors.NewGoCollector())

	serverMetrics.TotalRequestNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_grpc_request_number",
		Help: "The total number of grpc requests this server has received.",
	})
	serverMetrics.Registry.MustRegister(serverMetrics.TotalRequestNumber)

	serverMetrics.RequestDurations = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "A histogram of the grpc request durations in seconds.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})
	serverMetrics.Registry.MustRegister(serverMetrics.RequestDurations)

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(serverMetrics.Registry, promhttp.HandlerOpts{Registry: serverMetrics.Registry}))
		http.ListenAndServe(":8080", nil)
	}()

	return serverMetrics
}

func (s *ServerMetrics) AddRequestNum() {
	atomic.AddInt64(&s.RequestNum, 1)
}

func (s *ServerMetrics) GetRequestNum() int64 {
	return atomic.LoadInt64(&s.RequestNum)
}

func (s *ServerMetrics) ClearRequestNum() {
	atomic.StoreInt64(&s.RequestNum, 0)
}

type ServerPerformance struct {
	NumGoroutine      int     `json:"num_goroutine"`
	CpuPercent        float64 `json:"cpu_percent"`
	MemPercent        float64 `json:"mem_percent"`
	DiskPercent       float64 `json:"disk_percent"`
	RequestInDuration int64   `json:"request_in_duration"`
}

func (s *ServerPerformance) ToString() string {
	jsonData, err := json.Marshal(s)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return string(jsonData)
}

func ParseServerPerformance(s string) *ServerPerformance {
	var sp ServerPerformance
	err := json.Unmarshal([]byte(s), &sp)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return &sp
}
