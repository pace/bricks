// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pace/bricks/maintenance/log/hlog"
	"github.com/pace/bricks/maintenance/tracing/wire"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestIDHeader name of the header that can contain a request ID
const RequestIDHeader = "Request-Id"

// Handler returns a middleware that handles all of the logging aspects of
// any incoming http request. Optionally several path prefixes like "/health"
// can be provided to decrease log spamming. All url paths with these
// prefixes will not be logged to the standard output but still be available
// in the request specific Sink.
func Handler(silentPrefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return hlog.NewHandler(log.Logger)(
			handlerWithSink(silentPrefixes...)(
				hlog.AccessHandler(requestCompleted)(
					RequestIDHandler("req_id", RequestIDHeader)(next))))
	}
}

// requestCompleted logs all request related information once
// at the end of the request
var requestCompleted = func(r *http.Request, status, size int, duration time.Duration) {
	// log if the tracing id came from the wire
	_, err := wire.FromWire(r)
	var val string
	if err != nil {
		val = "new"
	} else {
		val = "wire"
	}

	hlog.FromRequest(r).Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status", status).
		Str("host", r.Host).
		Int("size", size).
		Dur("duration", duration).
		Str("ip", ProxyAwareRemote(r)).
		Str("referer", r.Header.Get("Referer")).
		Str("user_agent", r.Header.Get("User-Agent")).
		Str("span", val).
		Msg("Request Completed")
}

// ProxyAwareRemote return the most likely remote address
func ProxyAwareRemote(r *http.Request) string {
	// if we get the content via a proxy, try to extract the
	// ip from the usual headers
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() {
				continue // bad address, go to next
			}
			return ip
		}
	}
	// if no proxy header is present return the
	// regular remote address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Ctx(r.Context()).Warn().Err(err).Msg("failed to decode the remote address")
		return ""
	}
	return host
}

var noXid = errors.New("no xid")

func RequestIDHandler(fieldKey, headerName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var id xid.ID
			var err error

			// try extract of xid from header
			reqID := r.Header.Get(headerName)
			if reqID != "" {
				id, err = xid.FromString(reqID)
			} else {
				err = noXid
			}

			// try extract of xid context
			if err != nil {
				nid, ok := hlog.IDFromCtx(ctx)
				if !ok {
					// give up, generate a new xid
					id = xid.New()
				}
				id = nid
			}

			ctx = hlog.WithValue(ctx, id)
			r = r.WithContext(ctx)

			// log requests with request id
			log := zerolog.Ctx(ctx)
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str(fieldKey, id.String())
			})

			// return the request id as a header to the client
			w.Header().Set(headerName, id.String())
			next.ServeHTTP(w, r)
		})
	}
}
