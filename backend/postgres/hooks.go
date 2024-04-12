package postgres

import (
	"context"
	"math"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/pace/bricks/maintenance/log"
)

type QueryLogger struct{}

func (QueryLogger) BeforeQuery(_ context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return nil, nil
}

func (QueryLogger) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	q, qe := event.UnformattedQuery()
	if qe == nil {
		if !(cfg.LogRead || cfg.LogWrite) {
			return nil
		}
		// we can only and should only perfom the following check if we have the information availaible
		mode := determineQueryMode(string(q))
		if mode == readMode && !cfg.LogRead {
			return nil
		}
		if mode == writeMode && !cfg.LogWrite {
			return nil
		}

	}

	dur := float64(time.Since(event.StartTime)) / float64(time.Millisecond)

	// check if log context is given
	var logger *zerolog.Logger
	if ctx != nil {
		logger = log.Ctx(ctx)
	} else {
		logger = log.Logger()
	}

	// add general info
	le := logger.Debug().
		Float64("duration", dur).
		Str("sentry:category", "postgres")

	// add error or result set info
	if event.Err != nil {
		le = le.Err(event.Err)
	} else {
		le = le.Int("affected", event.Result.RowsAffected()).
			Int("rows", event.Result.RowsReturned())
	}

	if qe != nil {
		// this is only a display issue not a "real" issue
		le.Msgf("%v", qe)
	}
	le.Msg(string(q))

	return nil
}

type OpenTracingAdapter struct{}

func (OpenTracingAdapter) BeforeQuery(_ context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return nil, nil
}

func (OpenTracingAdapter) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	// start span with general info
	q, qe := event.UnformattedQuery()
	if qe != nil {
		// this is only a display issue not a "real" issue
		q = []byte(qe.Error())
	}

	span, _ := opentracing.StartSpanFromContext(event.DB.Context(), "sql: "+getQueryType(string(q)),
		opentracing.StartTime(event.StartTime))

	span.SetTag("db.system", "postgres")

	fields := []olog.Field{
		olog.String("query", string(q)),
	}

	// add error or result set info
	if event.Err != nil {
		fields = append(fields, olog.Error(event.Err))
	} else {
		fields = append(fields,
			olog.Int("affected", event.Result.RowsAffected()),
			olog.Int("rows", event.Result.RowsReturned()))
	}

	span.LogFields(fields...)
	span.Finish()

	return nil
}

type MetricsAdapter struct {
	opts *pg.Options
}

func (MetricsAdapter) BeforeQuery(_ context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return nil, nil
}

func (m MetricsAdapter) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	dur := float64(time.Since(event.StartTime)) / float64(time.Millisecond)
	labels := prometheus.Labels{
		"database": m.opts.Addr + "/" + m.opts.Database,
	}

	metricQueryTotal.With(labels).Inc()

	if event.Err != nil {
		metricQueryFailed.With(labels).Inc()
	} else {
		r := event.Result
		metricQueryRowsTotal.With(labels).Add(float64(r.RowsReturned()))
		metricQueryAffectedTotal.With(labels).Add(math.Max(0, float64(r.RowsAffected())))
	}
	metricQueryDurationSeconds.With(labels).Observe(dur)

	return nil
}
