package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ClientMetrics struct {
	Registry            *prometheus.Registry
	TotalResponseNumber prometheus.Gauge
	RequestDurations    prometheus.Summary
	RequestDestination  prometheus.GaugeVec
}

func NewClientMetrics() *ClientMetrics {
	clientMetrics := &ClientMetrics{}

	clientMetrics.Registry = prometheus.NewRegistry()

	clientMetrics.Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	clientMetrics.Registry.MustRegister(collectors.NewGoCollector())

	clientMetrics.TotalResponseNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_response_number",
		Help: "The total number of responses this client has received.",
	})
	clientMetrics.Registry.MustRegister(clientMetrics.TotalResponseNumber)

	clientMetrics.RequestDurations = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "request_duration_seconds",
		Help: "A summary of the request durations in seconds.",
		Objectives: map[float64]float64{
			0.1:  0.05,
			0.5:  0.05,  // 第50个百分位数，最大绝对误差为0.05。
			0.9:  0.01,  // 第90个百分位数，最大绝对误差为0.01。
			0.99: 0.001, // 第99个百分位数，最大绝对误差为0.001。
		},
	},
	)
	clientMetrics.Registry.MustRegister(clientMetrics.RequestDurations)

	clientMetrics.RequestDestination = *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "request_destination",
		Help: "request destinations of this client has sent",
	},
		[]string{"port"})
	clientMetrics.Registry.MustRegister(clientMetrics.RequestDestination)

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(clientMetrics.Registry, promhttp.HandlerOpts{Registry: clientMetrics.Registry}))
		http.ListenAndServe(":8080", nil)
	}()

	return clientMetrics
}
