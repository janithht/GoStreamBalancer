package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var CustomRegistry = prometheus.NewRegistry()

var (
	TotalRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loadbalancer_requests_total",
		Help: "Total number of requests processed by the load balancer.",
	}, []string{"upstream"})

	SuccessfulRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loadbalancer_requests_successful",
		Help: "Total number of successful requests processed by the load balancer.",
	}, []string{"upstream"})

	RequestErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loadbalancer_request_errors_total",
		Help: "Total number of request errors by type and upstream.",
	}, []string{"status_code", "upstream"})

	UpstreamConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "loadbalancer_upstream_connections",
		Help: "Current number of active connections to upstream servers.",
	}, []string{"upstream"})

	RateLimitHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loadbalancer_rate_limit_hits_total",
		Help: "Total number of times rate limits were hit per upstream.",
	}, []string{"upstream"})

	RequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "loadbalancer_request_latency_seconds",
		Help:    "Histogram of latencies for incoming requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"upstream"})

	ResponseTimes = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "loadbalancer_response_times_milliseconds",
		Help:    "Histogram of response times of the load balancer in milliseconds",
		Buckets: prometheus.LinearBuckets(10, 10, 10), // Start at 10ms with 10ms increments
	})
	TCPActiveConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcp_loadbalancer_active_connections",
		Help: "Current number of active TCP connections to upstream servers.",
	}, []string{"upstream"})

	TCPThroughput = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tcp_loadbalancer_throughput_total",
		Help: "Total number of TCP requests processed by the load balancer.",
	}, []string{"upstream"})

	TCPBytesTransferred = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tcp_loadbalancer_bytes_transferred_total",
		Help: "Total number of bytes transferred by the load balancer.",
	}, []string{"upstream"})
)

func init() {
	CustomRegistry.MustRegister(TotalRequests, SuccessfulRequests, RequestErrors, UpstreamConnections, RateLimitHits, RequestLatency, ResponseTimes, TCPActiveConnections, TCPThroughput, TCPBytesTransferred)
}

func RecordRequest(upstream string) {
	TotalRequests.WithLabelValues(upstream).Inc()
}

func RecordSuccess(upstream string) {
	SuccessfulRequests.WithLabelValues(upstream).Inc()
}

func RecordError(statusCode, upstream string) {
	RequestErrors.WithLabelValues(statusCode, upstream).Inc()
}

func SetConnections(upstream string, count float64) {
	UpstreamConnections.WithLabelValues(upstream).Set(count)
}

func RecordRateLimitHit(upstream string) {
	RateLimitHits.WithLabelValues(upstream).Inc()
}

func RecordTCPRequest(upstream string) {
	TCPThroughput.WithLabelValues(upstream).Inc()
}

func SetTCPConnections(upstream string, count float64) {
	TCPActiveConnections.WithLabelValues(upstream).Set(count)
}

func RecordThroughput(upstream string, bytes float64) {
	TCPBytesTransferred.WithLabelValues(upstream).Add(bytes)
}
