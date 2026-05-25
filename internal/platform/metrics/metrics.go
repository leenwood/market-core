package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	Registry            *prometheus.Registry
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	OutboundTotal       *prometheus.CounterVec
	QueueLag            *prometheus.GaugeVec
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	httpTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_http_requests_total",
		Help: "Total HTTP requests by method, path, and status",
	}, []string{"method", "path", "status"})

	httpDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "platform_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"method", "path"})

	outbound := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_http_outbound_requests_total",
		Help: "Total outbound HTTP requests by target, method, and status",
	}, []string{"target", "method", "status"})

	queueLag := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_queue_lag_messages",
		Help: "Kafka consumer queue lag by topic",
	}, []string{"topic"})

	reg.MustRegister(httpTotal, httpDuration, outbound, queueLag)

	return &Metrics{
		Registry:            reg,
		HTTPRequestsTotal:   httpTotal,
		HTTPRequestDuration: httpDuration,
		OutboundTotal:       outbound,
		QueueLag:            queueLag,
	}
}
