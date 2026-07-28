package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veil_http_requests_total",
			Help: "Total HTTP requests served by Veil services.",
		},
		[]string{"service", "method", "route", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veil_http_request_duration_seconds",
			Help:    "HTTP request latency for Veil services.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "route"},
	)
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// RegisterMetrics attaches GET /metrics (Prometheus exposition) to mux.
func RegisterMetrics(mux *http.ServeMux) {
	mux.Handle("GET /metrics", promhttp.Handler())
}

// InstrumentHTTP records RED metrics for requests handled by next.
func InstrumentHTTP(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		route := r.URL.Path
		status := strconv.Itoa(rec.status)
		httpRequestsTotal.WithLabelValues(service, r.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(service, r.Method, route).Observe(time.Since(start).Seconds())
	})
}

// WrapHandler adds /metrics and instrumentation around an existing handler.
func WrapHandler(service string, handler http.Handler) http.Handler {
	mux := http.NewServeMux()
	RegisterMetrics(mux)
	mux.Handle("/", handler)
	return ChainHTTP(service, mux)
}
