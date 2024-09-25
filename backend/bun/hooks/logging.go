package hooks

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"

	"github.com/pace/bricks/maintenance/log"
)

type queryMode int

const (
	readMode  queryMode = iota
	writeMode queryMode = iota
)

type LoggingHook struct {
	logReadQueries  bool
	logWriteQueries bool
}

func NewLoggingHook(logRead bool, logWrite bool) *LoggingHook {
	return &LoggingHook{
		logReadQueries:  logRead,
		logWriteQueries: logWrite,
	}
}

func (h *LoggingHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *LoggingHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	// we can only and should only perfom the following check if we have the information availaible
	mode := determineQueryMode(event.Query)

	if mode == readMode && !h.logReadQueries {
		return
	}

	if mode == writeMode && !h.logWriteQueries {
		return
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
	logEvent := logger.Debug().
		Float64("duration", dur).
		Str("sentry:category", "postgres")

	// add error or result set info
	if event.Err != nil {
		logEvent = logEvent.Err(event.Err)
	} else {
		rowsAffected, err := event.Result.RowsAffected()
		if err == nil {
			logEvent = logEvent.Int64("affected", rowsAffected)
		}
	}

	logEvent.Msg(event.Query)
}

// determineQueryMode is a poorman's attempt at checking whether the query is a read or write to the database.
// Feel free to improve.
func determineQueryMode(qry string) queryMode {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(qry)), "select") {
		return readMode
	}
	return writeMode
}
