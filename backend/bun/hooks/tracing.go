package hooks

import (
	"context"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/uptrace/bun"
)

var (
	reQueryType        = regexp.MustCompile(`(\s)`)
	reQueryTypeCleanup = regexp.MustCompile(`(?m)(\s+|\n)`)
)

type TracingHook struct{}

func (h *TracingHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *TracingHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	span, _ := opentracing.StartSpanFromContext(ctx, "sql: "+getQueryType(event.Query),
		opentracing.StartTime(event.StartTime))

	span.SetTag("db.system", "postgres")

	fields := []olog.Field{
		olog.String("query", event.Query),
	}

	// add error or result set info
	if event.Err != nil {
		fields = append(fields, olog.Error(event.Err))
	} else {
		rowsAffected, err := event.Result.RowsAffected()
		if err == nil {
			fields = append(fields, olog.Int64("affected", rowsAffected))
		}
	}

	span.LogFields(fields...)
	span.Finish()
}

func getQueryType(s string) string {
	s = reQueryTypeCleanup.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	p := reQueryType.FindStringIndex(s)
	if len(p) > 0 {
		return strings.ToUpper(s[:p[0]])
	}

	return strings.ToUpper(s)
}
