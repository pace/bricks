// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/20 by Vincent Landgraf

package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	raven "github.com/getsentry/raven-go"
	"lab.jamit.de/pace/go-microservice/http/jsonapi/runtime"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
	"lab.jamit.de/pace/go-microservice/oauth2"
)

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

// Handler implements a panic recovering middleware
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return &recoveryHandler{next} }
}

// HandleRequest should be called with defer to recover panics in request handlers
func HandleRequest(handlerName string, w http.ResponseWriter, r *http.Request) {
	if rp := recover(); rp != nil {
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
	userId, _ := oauth2.UserID(ctx)
	packet.Interfaces = append(packet.Interfaces, &raven.User{ID: userId, IP: log.ProxyAwareRemote(r)})
	appendInfoFromContext(ctx, packet, handlerName, log.RequestID(r))

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
		userId, _ := oauth2.UserID(ctx)
		packet.Interfaces = append(packet.Interfaces, &raven.User{ID: userId})
		appendInfoFromContext(ctx, packet, handlerName, requestIDFromContext(ctx))

		raven.Capture(packet, nil)
	}
}

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// Adds extra data to an error before reporting to Sentry
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

	// extraxt ErrWithExtra info and append it to the packet
	if ee, ok := rp.(raven.ErrWithExtra); ok {
		for k, v := range ee.ExtraInfo() {
			packet.Extra[k] = v
		}
	}

	return packet
}

func appendInfoFromContext(ctx context.Context, packet *raven.Packet, handlerName, reqID string) {
	packet.Extra["req_id"] = reqID
	packet.Extra["handler"] = handlerName
	if clientID, ok := oauth2.ClientID(ctx); ok {
		packet.Extra["oauth2_client_id"] = clientID
	}
	if scopes := oauth2.Scopes(ctx); len(scopes) > 0 {
		packet.Extra["oauth2_scopes"] = scopes
	}

	packet.Extra["microservice"] = os.Getenv("JAEGER_SERVICE_NAME")
}

func requestIDFromContext(ctx context.Context) string {
	// Note: a hack to get to the RequestID using a dummy request
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(fmt.Errorf("New request dummy creation failed: %v", err))
	}
	return log.RequestID(r.WithContext(ctx))
}
