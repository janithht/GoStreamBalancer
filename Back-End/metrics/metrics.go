package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var CustomRegistry = prometheus.NewRegistry()

var (
	totalRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadbalancer_requests_total",
		Help: "Total number of requests processed by the load balancer.",
	})
	requestErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loadbalancer_request_errors_total",
		Help: "Total number of request errors by type.",
	}, []string{"status_code"})
	upstreamConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "loadbalancer_upstream_connections",
		Help: "Current number of active connections to upstream servers.",
	}, []string{"server_name"})
	rateLimitHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadbalancer_rate_limit_hits_total",
		Help: "Total number of times rate limits were hit.",
	})
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "loadbalancer_request_latency_seconds",
		Help:    "Histogram of latencies for incoming requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"endpoint"})
	ResponseTimes = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "loadbalancer_response_times",
		Help:    "Histogram of response times of the load balancer in milliseconds",
		Buckets: prometheus.LinearBuckets(10, 10, 10), // Adjust buckets as needed
	})
)

func init() {
	CustomRegistry.MustRegister(totalRequests, requestErrors, upstreamConnections, rateLimitHits, requestLatency, ResponseTimes)
}

func RecordRequest() {
	totalRequests.Inc()
}

func RecordError(statusCode string) {
	requestErrors.With(prometheus.Labels{"status_code": statusCode}).Inc()
}

func SetConnections(serverName string, count float64) {
	upstreamConnections.With(prometheus.Labels{"server_name": serverName}).Set(count)
}

func RecordRateLimitHit() {
	rateLimitHits.Inc()
}
