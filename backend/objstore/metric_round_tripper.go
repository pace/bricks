package objstore

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	paceObjStoreTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_objstore_req_total",
			Help: "Collects stats about the number of object storage requests made",
		},
		[]string{"method", "bucket"},
	)
	paceObjStoreFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_objstore_req_failed",
			Help: "Collects stats about the number of object storage requests counterFailed",
		},
		[]string{"method", "bucket"},
	)
	paceObjStoreDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_objstore_req_duration_seconds",
			Help:    "Collect performance metrics for each method & bucket",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"method", "bucket"},
	)
)

type metricRoundTripper struct {
	transport http.RoundTripper
	endpoint  string
}

func (m *metricRoundTripper) Transport() http.RoundTripper {
	return m.transport
}

func (m *metricRoundTripper) SetTransport(rt http.RoundTripper) {
	m.transport = rt
}

func (m *metricRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	labels := prometheus.Labels{
		"method": req.Method,
		"bucket": m.endpoint,
	}

	start := time.Now()
	resp, err := m.Transport().RoundTrip(req)
	dur := time.Since(start)

	// total
	paceObjStoreTotal.With(labels).Inc()

	// duration
	measurable := err != nil
	if measurable {
		// no need to measure timeouts or transport issues
		paceObjStoreDurationSeconds.With(labels).Observe(dur.Seconds())
	}

	// failure
	failed := err != nil || m.determineFailure(resp.StatusCode)
	if failed {
		// count transport issues and by status code
		paceObjStoreFailed.With(labels).Inc()
	}

	return resp, err
}

// determineFailure determines whether the response code is considered failed or not.
func (m *metricRoundTripper) determineFailure(code int) bool {
	return !(200 <= code && code < 400)
}
