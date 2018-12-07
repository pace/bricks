package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	paceHttpInFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pace_http_in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	paceHttpCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_http_request_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	paceHttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_http_request_duration_milliseconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method"},
	)

	// responseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	paceHttpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_http_request_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{100, 200, 500, 900, 1500},
		},
		[]string{"code", "method"},
	)
)

func init() {
	// Register all of the metrics in the standard registry.
	prometheus.MustRegister(paceHttpInFlightGauge, paceHttpCounter, paceHttpDuration, paceHttpResponseSize)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paceHttpInFlightGauge.Inc()
		startTime := time.Now()
		srw := statusWriter{ResponseWriter: w}
		next.ServeHTTP(&srw, r)
		dur := float64(time.Since(startTime)) / float64(time.Millisecond)
		labels := prometheus.Labels{
			"code":   strconv.Itoa(srw.status),
			"method": r.Method,
		}
		paceHttpCounter.With(labels).Inc()
		paceHttpDuration.With(labels).Observe(dur)
		paceHttpResponseSize.With(labels).Observe(float64(srw.length))
		paceHttpInFlightGauge.Dec()
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
