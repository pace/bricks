// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/20 by Vincent Landgraf

package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/maintenance/errors/raven"
	"github.com/pace/bricks/maintenance/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	paceHTTPPanicCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pace_http_panic_total",
		Help: "A counter for panics intercepted while handling a request",
	})
)

func init() {
	prometheus.MustRegister(paceHTTPPanicCounter)
}

// PanicWrap wraps a panic for HandleRequest
type PanicWrap struct {
	err interface{}
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
		return v.(*http.Request)
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

// Handler implements a panic recovering middleware
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		next = &contextHandler{next: next}
		return &recoveryHandler{next: next}
	}
}

// HandleRequest should be called with defer to recover panics in request handlers
func HandleRequest(handlerName string, w http.ResponseWriter, r *http.Request) {
	if rp := recover(); rp != nil {
		paceHTTPPanicCounter.Inc()
		HandleError(&PanicWrap{rp}, handlerName, w, r)
	}
}

// HandleError reports the passed error to sentry
func HandleError(rp interface{}, handlerName string, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pw, ok := rp.(*PanicWrap)
	if ok {
		log.Ctx(ctx).Error().Str("handler", handlerName).Msgf("Panic: %v", pw.err)
		rp = pw.err // unwrap error
	} else {
		log.Ctx(ctx).Error().Str("handler", handlerName).Msgf("Error: %v", rp)
	}
	log.Stack(ctx)

	sentryEvent{ctx, r, rp, 1, handlerName}.Send()

	runtime.WriteError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
}

// Handle logs the given error and reports it to sentry.
func Handle(ctx context.Context, rp interface{}) {
	pw, ok := rp.(*PanicWrap)
	if ok {
		log.Ctx(ctx).Error().Msgf("Panic: %v", pw.err)
		rp = pw.err // unwrap error
	} else {
		log.Ctx(ctx).Error().Msgf("Error: %v", rp)
	}
	log.Stack(ctx)

	sentryEvent{ctx, nil, rp, 1, ""}.Send()
}

// HandleWithCtx should be called with defer to recover panics in goroutines
func HandleWithCtx(ctx context.Context, handlerName string) {
	if rp := recover(); rp != nil {
		log.Ctx(ctx).Error().Str("handler", handlerName).Msgf("Panic: %v", rp)
		log.Stack(ctx)

		sentryEvent{ctx, nil, rp, 2, handlerName}.Send()
	}
}

func HandleErrorNoStack(ctx context.Context, err error) {
	log.Ctx(ctx).Info().Msgf("Received error, will not handle further: %v", err)
}

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// WrapWithExtra adds extra data to an error before reporting to Sentry
func WrapWithExtra(err error, extraInfo map[string]interface{}) error {
	return raven.WrapWithExtra(err, extraInfo)
}

type sentryEvent struct {
	ctx         context.Context
	req         *http.Request // optional
	r           interface{}
	level       int
	handlerName string
}

func (e sentryEvent) Send() {
	_, errCh := raven.Capture(e.build(), nil)
	<-errCh // ensure the message get send even if the main goroutine is about to stop
}

func (e sentryEvent) build() *raven.Packet {
	ctx, r, rp, handlerName := e.ctx, e.req, e.r, e.handlerName

	// get request from context if available
	if r == nil {
		r = requestFromContext(ctx)
	}

	rvalStr := fmt.Sprint(rp)
	var packet *raven.Packet

	if err, ok := rp.(error); ok {
		stack := raven.GetOrNewStacktrace(err, 2+e.level, 3, nil)
		packet = raven.NewPacket(rvalStr, raven.NewException(err, stack))
	} else {
		stack := raven.NewStacktrace(2+e.level, 3, nil)
		packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), stack))
	}

	// extract ErrWithExtra info and append it to the packet
	if ee, ok := rp.(raven.ErrWithExtra); ok {
		for k, v := range ee.ExtraInfo() {
			packet.Extra[k] = v
		}
	}

	// add user
	userID, ok := oauth2.UserID(ctx)
	user := raven.User{ID: userID}
	if r != nil {
		user.IP = log.ProxyAwareRemote(r)
	}
	packet.Interfaces = append(packet.Interfaces, &user)
	if ok {
		packet.Tags = append(packet.Tags, raven.Tag{Key: "user_id", Value: userID})
	}

	// from context
	if reqID := log.RequestIDFromContext(ctx); reqID != "" {
		packet.Extra["req_id"] = reqID
		packet.Tags = append(packet.Tags, raven.Tag{Key: "req_id", Value: reqID})
	}
	if traceID := log.TraceIDFromContext(ctx); traceID != "" {
		packet.Extra["uber_trace_id"] = traceID
		packet.Tags = append(packet.Tags, raven.Tag{Key: "trace_id", Value: traceID})
	}
	packet.Extra["handler"] = handlerName
	if clientID, ok := oauth2.ClientID(ctx); ok {
		packet.Extra["oauth2_client_id"] = clientID
	}
	if scopes := oauth2.Scopes(ctx); len(scopes) > 0 {
		packet.Extra["oauth2_scopes"] = scopes
	}

	// from request
	if r != nil {
		packet.Interfaces = append(packet.Interfaces, raven.NewHttp(r))
	}

	// from env
	packet.Extra["microservice"] = os.Getenv("JAEGER_SERVICE_NAME")

	// add breadcrumbs
	packet.Breadcrumbs = getBreadcrumbs(ctx)

	return packet
}

// getBreadcrumbs takes a context and tries to extract the logs from it if it
// holds a log.Sink. If that's the case, the logs will all be translated
// to valid sentry breadcrumbs if possible. In case of a failure, the
// breadcrumbs will be dropped and a warning will be logged.
func getBreadcrumbs(ctx context.Context) []*raven.Breadcrumb {
	sink, ok := log.SinkFromContext(ctx)
	if !ok {
		return nil
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(sink.ToJSON(), &data); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to prepare sentry message")
		return nil
	}

	result := make([]*raven.Breadcrumb, len(data))
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

func createBreadcrumb(data map[string]interface{}) (*raven.Breadcrumb, error) {
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

	return &raven.Breadcrumb{
		Category:  category,
		Level:     level,
		Message:   message,
		Timestamp: time.Unix(),
		Type:      typ,
		Data:      data,
	}, nil
}

// translateZerologLevelToSentryLevel takes in a zerolog.Level as string
// and returns the equivalent sentry breadcrumb level. If the given level
// can't be parsed to a valid zerolog.Level an error is returned.
func translateZerologLevelToSentryLevel(l string) (string, error) {
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
