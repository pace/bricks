package hooks

import (
	"context"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uptrace/bun"
)

var (
	MetricQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_total",
			Help: "Collects stats about the number of postgres queries made",
		},
		[]string{"database"},
	)
	MetricQueryFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_failed",
			Help: "Collects stats about the number of postgres queries failed",
		},
		[]string{"database"},
	)
	MetricQueryDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_postgres_query_duration_seconds",
			Help:    "Collect performance metrics for each postgres query",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"database"},
	)
	MetricQueryAffectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_affected_total",
			Help: "Collects stats about the number of rows affected by a postgres query",
		},
		[]string{"database"},
	)
)

type MetricsHook struct {
	addr     string
	database string
}

func NewMetricsHook(addr string, database string) *MetricsHook {
	return &MetricsHook{
		addr:     addr,
		database: database,
	}
}

func (h *MetricsHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *MetricsHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	dur := float64(time.Since(event.StartTime)) / float64(time.Millisecond)

	labels := prometheus.Labels{
		"database": h.addr + "/" + h.database,
	}

	MetricQueryTotal.With(labels).Inc()

	if event.Err != nil {
		MetricQueryFailed.With(labels).Inc()
	} else {
		r := event.Result
		rowsAffected, err := r.RowsAffected()
		if err == nil {
			MetricQueryAffectedTotal.With(labels).Add(math.Max(0, float64(rowsAffected)))
		}
	}

	MetricQueryDurationSeconds.With(labels).Observe(dur)
}
