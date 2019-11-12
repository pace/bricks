// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/20 by Vincent Landgraf

package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/getsentry/raven-go"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/http/security/oauth2"
	"github.com/pace/bricks/maintenance/log"
	"github.com/prometheus/client_golang/prometheus"
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
	counter prometheus.Gauge
	next    http.Handler
}

func (h *recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer HandleRequest("rootHandler", w, r)
	h.next.ServeHTTP(w, r)
}

// Handler implements a panic recovering middleware
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return &recoveryHandler{next: next} }
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

	packet := newPacket(rp)

	// append http info
	packet.Interfaces = append(packet.Interfaces, raven.NewHttp(r))

	// append additional info
	userID, ok := oauth2.UserID(ctx)
	packet.Interfaces = append(packet.Interfaces, &raven.User{ID: userID, IP: log.ProxyAwareRemote(r)})
	if ok {
		packet.Tags = append(packet.Tags, raven.Tag{Key: "user_id", Value: userID})
	}
	appendInfoFromContext(ctx, packet, handlerName)

	raven.Capture(packet, nil)

	runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
}

// HandleWithCtx should be called with defer to recover panics in goroutines
func HandleWithCtx(ctx context.Context, handlerName string) {
	if rp := recover(); rp != nil {
		log.Ctx(ctx).Error().Str("handler", handlerName).Msgf("Panic: %v", rp)
		log.Stack(ctx)

		packet := newPacket(rp)

		// append additional info
		userID, ok := oauth2.UserID(ctx)
		packet.Interfaces = append(packet.Interfaces, &raven.User{ID: userID})
		if ok {
			packet.Tags = append(packet.Tags, raven.Tag{Key: "user_id", Value: userID})
		}
		appendInfoFromContext(ctx, packet, handlerName)

		raven.Capture(packet, nil)
	}
}

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// WrapWithExtra adds extra data to an error before reporting to Sentry
func WrapWithExtra(err error, extraInfo map[string]interface{}) error {
	return raven.WrapWithExtra(err, extraInfo)
}

func newPacket(rp interface{}) *raven.Packet {
	rvalStr := fmt.Sprint(rp)
	var packet *raven.Packet

	if err, ok := rp.(error); ok {
		packet = raven.NewPacket(rvalStr, raven.NewException(err, raven.GetOrNewStacktrace(err, 2, 3, nil)))
	} else {
		packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
	}

	// extract ErrWithExtra info and append it to the packet
	if ee, ok := rp.(raven.ErrWithExtra); ok {
		for k, v := range ee.ExtraInfo() {
			packet.Extra[k] = v
		}
	}

	return packet
}

func appendInfoFromContext(ctx context.Context, packet *raven.Packet, handlerName string) {
	if reqID := log.RequestIDFromContext(ctx); reqID != "" {
		packet.Extra["req_id"] = reqID
		packet.Tags = append(packet.Tags, raven.Tag{Key: "req_id", Value: reqID})
	}
	packet.Extra["handler"] = handlerName
	if clientID, ok := oauth2.ClientID(ctx); ok {
		packet.Extra["oauth2_client_id"] = clientID
	}
	if scopes := oauth2.Scopes(ctx); len(scopes) > 0 {
		packet.Extra["oauth2_scopes"] = scopes
	}

	packet.Extra["microservice"] = os.Getenv("JAEGER_SERVICE_NAME")
}
