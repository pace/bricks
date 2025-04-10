// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/http/oauth2"
	_ "github.com/pace/bricks/internal/sentry"
	"github.com/pace/bricks/maintenance/log"
)

var paceHTTPPanicCounter = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "pace_http_panic_total",
	Help: "A counter for panics intercepted while handling a request",
})

var DefaultClient *sentry.Client

func init() {
	prometheus.MustRegister(paceHTTPPanicCounter)
}

// ExtraError wraps an error and adds extra information to it.
type ExtraError struct {
	err   error
	extra map[string]any
}

// NewExtraError creates a new ExtraError with the given error and extra.
func NewExtraError(err error, extra map[string]any) ExtraError {
	return ExtraError{
		err:   err,
		extra: extra,
	}
}

// Error implements the error interface.
func (e ExtraError) Error() string {
	return e.err.Error()
}

// PanicError is a wrapper for panics that occur in the code.
type PanicError struct {
	err any
}

// NewPanicError creates a new PanicError with the given error.
func NewPanicError(err any) PanicError {
	return PanicError{err: err}
}

// Error implements the error interface.
func (p PanicError) Error() string {
	return fmt.Sprintf("%v", p.err)
}

type recoveryHandler struct {
	next http.Handler
}

func (h *recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer HandleRequest("rootHandler", w, r)
	h.next.ServeHTTP(w, r)
}

type ctxKey struct{}

var reqKey = ctxKey{}

func contextWithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, reqKey, r)
}

func requestFromContext(ctx context.Context) *http.Request {
	if v := ctx.Value(reqKey); v != nil {
		if out, ok := v.(*http.Request); ok {
			return out
		}
	}

	return nil
}

// ContextTransfer copies error handling related information from one context to
// another.
func ContextTransfer(ctx, targetCtx context.Context) context.Context {
	if r := requestFromContext(ctx); r != nil {
		return contextWithRequest(targetCtx, r)
	}

	return targetCtx
}

type contextHandler struct {
	next http.Handler
}

func (h *contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For error handling we want to log information about the request under
	// which the error happened. But in some cases we only have a context,
	// because unlike the context the request is not passed down. To make the
	// request available for error handling we add it to the context here.
	h.next.ServeHTTP(w, r.WithContext(contextWithRequest(r.Context(), r)))
}

// Handler implements a panic recovering middleware.
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		next = &contextHandler{next: next}
		return &recoveryHandler{next: next}
	}
}

// HandleRequest should be called with defer to recover panics in request handlers.
func HandleRequest(handlerName string, w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		paceHTTPPanicCounter.Inc()
		HandleError(NewPanicError(rec), handlerName, w, r)
	}
}

// HandleError reports the passed error to sentry.
func HandleError(err error, handlerName string, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handle(ctx, err, handlerName)

	runtime.WriteError(w, http.StatusInternalServerError, errors.New("internal Server Error"))
}

// Handle logs the given error and reports it to sentry.
func Handle(ctx context.Context, err error) {
	handle(ctx, err, "")
}

func handle(ctx context.Context, err error, handlerName string) {
	l := log.Ctx(ctx).Error().Err(err)

	if handlerName != "" {
		l = l.Str("handler", handlerName)
	}

	var p PanicError

	if errors.As(err, &p) {
		l.Msg("Panic")
	} else {
		l.Msg("Error")
	}

	log.Stack(ctx)

	sentry.CaptureEvent(getEvent(ctx, nil, err, 1, handlerName))
}

func getEvent(ctx context.Context, r *http.Request, err error, level int, handlerName string) *sentry.Event {
	// get request from context if available
	if r == nil {
		r = requestFromContext(ctx)
	}

	event := sentry.NewEvent()

	event.SetException(err, 2+level)

	// add user
	userID, ok := oauth2.UserID(ctx)
	if ok {
		event.User.ID = userID
	}

	if r != nil {
		event.User.IPAddress = log.ProxyAwareRemote(r)
	}

	// from context
	if reqID := log.RequestIDFromContext(ctx); reqID != "" {
		event.Extra["req_id"] = reqID
		event.Tags["req_id"] = reqID
	}

	if traceID := log.TraceIDFromContext(ctx); traceID != "" {
		event.Extra["uber_trace_id"] = traceID
		event.Tags["trace_id"] = traceID
	}

	event.Extra["handler"] = handlerName

	if clientID, ok := oauth2.ClientID(ctx); ok {
		event.Extra["oauth2_client_id"] = clientID
	}

	if scopes := oauth2.Scopes(ctx); len(scopes) > 0 {
		event.Extra["oauth2_scopes"] = scopes
	}

	// from env
	event.Extra["microservice"] = os.Getenv("JAEGER_SERVICE_NAME")

	// add breadcrumbs
	event.Breadcrumbs = getBreadcrumbs(ctx)

	return event
}

// HandleWithCtx should be called with defer to recover panics in goroutines.
func HandleWithCtx(ctx context.Context, handlerName string) {
	if r := recover(); r != nil {
		log.Ctx(ctx).Error().Str("handler", handlerName).Msgf("Panic: %v", r)
		log.Stack(ctx)

		sentry.CaptureEvent(getEvent(ctx, nil, NewPanicError(r), 2, handlerName))
	}
}

func HandleErrorNoStack(ctx context.Context, err error) {
	log.Ctx(ctx).Info().Msgf("Received error, will not handle further: %v", err)
}

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// WrapWithExtra adds extra data to an error before reporting to Sentry.
func WrapWithExtra(err error, extraInfo map[string]any) error {
	return NewExtraError(err, extraInfo)
}

// getBreadcrumbs takes a context and tries to extract the logs from it if it
// holds a log.Sink. If that's the case, the logs will all be translated
// to valid sentry breadcrumbs if possible. In case of a failure, the
// breadcrumbs will be dropped and a warning will be logged.
func getBreadcrumbs(ctx context.Context) []*sentry.Breadcrumb {
	sink, ok := log.SinkFromContext(ctx)
	if !ok {
		return nil
	}

	var data []map[string]any
	if err := json.Unmarshal(sink.ToJSON(), &data); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to prepare sentry message")
		return nil
	}

	result := make([]*sentry.Breadcrumb, len(data))

	for i, d := range data {
		crumb, err := createBreadcrumb(d)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to create sentry breadcrumb")
			return nil
		}

		result[i] = crumb
	}

	return result
}

func createBreadcrumb(data map[string]any) (*sentry.Breadcrumb, error) {
	// remove the request id if it can still be found in the logs
	delete(data, "req_id")

	timeRaw, ok := data["time"].(string)
	if !ok {
		return nil, errors.New(`cannot parse "time"`)
	}

	delete(data, "time")

	time, err := time.Parse(time.RFC3339, timeRaw)
	if err != nil {
		return nil, fmt.Errorf(`cannot parse "time": %w`, err)
	}

	levelRaw, ok := data["level"].(string)
	if !ok {
		return nil, errors.New(`cannot parse "level"`)
	}

	delete(data, "level")

	level, err := translateZerologLevelToSentryLevel(levelRaw)
	if err != nil {
		return nil, fmt.Errorf(`cannot parse "level": %w`, err)
	}

	message, ok := data["message"].(string)
	if !ok {
		return nil, errors.New(`cannot parse "message"`)
	}

	delete(data, "message")

	categoryRaw, ok := data["sentry:category"]
	if !ok {
		categoryRaw = ""
	}

	delete(data, "sentry:category")

	category, ok := categoryRaw.(string)
	if !ok {
		return nil, errors.New(`cannot parse "category"`)
	}

	typRaw, ok := data["sentry:type"]
	if !ok {
		typRaw = ""
	}

	delete(data, "sentry:type")

	typ, ok := typRaw.(string)
	if !ok {
		return nil, errors.New(`cannot parse "type"`)
	}

	if typ == "" && level == "fatal" {
		typ = "error"
	}

	return &sentry.Breadcrumb{
		Category:  category,
		Level:     level,
		Message:   message,
		Timestamp: time,
		Type:      typ,
		Data:      data,
	}, nil
}

// translateZerologLevelToSentryLevel takes in a zerolog.Level as string
// and returns the equivalent sentry breadcrumb level. If the given level
// can't be parsed to a valid zerolog.Level an error is returned.
func translateZerologLevelToSentryLevel(l string) (sentry.Level, error) {
	level, err := zerolog.ParseLevel(l)
	if err != nil {
		return "", err
	}

	switch level {
	case zerolog.TraceLevel, zerolog.InfoLevel:
		return "info", nil
	case zerolog.DebugLevel:
		return "debug", nil
	case zerolog.WarnLevel:
		return "warning", nil
	case zerolog.ErrorLevel:
		return "error", nil
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return "fatal", nil
	case zerolog.NoLevel:
		fallthrough
	default:
		return "", nil
	}
}
