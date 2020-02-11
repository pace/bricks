package testmetrics

import (
	"net/http/httptest"
	"testing"

	"github.com/pace/bricks/http"
	"github.com/pace/bricks/maintenance/metric"
	"github.com/stretchr/testify/assert"
)

type MetricsSuite struct {
	metrics []string
	t       *testing.T
	name    string
}

func Setup(t *testing.T, name string, metrics ...string) *MetricsSuite {
	return &MetricsSuite{
		name:    name,
		metrics: metrics,
		t:       t,
	}
}

func (s *MetricsSuite) Run() {
	s.t.Run(s.name, func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", http.RouteMetrics, nil)
		metric.Handler().ServeHTTP(rec, req)

		body := rec.Body.String()
		for _, expected := range s.metrics {
			assert.Contains(t, body, expected, "metric '%s' missing in %s", expected, http.RouteMetrics)
		}
	})
}
