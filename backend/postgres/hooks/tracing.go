// Copyright © 2024 by PACE Telematics GmbH. All rights reserved.

package hooks

import (
	"context"
	"regexp"
	"strings"

	"github.com/getsentry/sentry-go"
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
	span := sentry.StartSpan(ctx, "db.sql.query", sentry.WithDescription(getQueryType(event.Query)))
	defer span.Finish()

	span.StartTime = event.StartTime

	span.SetTag("db.system", "postgres")

	span.SetData("query", event.Query)

	// add error or result set info
	if event.Err != nil {
		span.SetData("error", event.Err)
	} else if event.Result != nil {
		rowsAffected, err := event.Result.RowsAffected()
		if err == nil {
			span.SetData("affected", rowsAffected)
		}
	}
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
