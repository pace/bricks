// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// Handler is a middleware that handles all of the logging aspects of
// any incoming http request
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return hlog.NewHandler(log.Logger)(
			hlog.AccessHandler(requestCompleted)(
				hlog.RequestIDHandler("req_id", "Request-Id")(next)))
	}
}

// requestCompleted logs all request related information once
// at the end of the request
var requestCompleted = func(r *http.Request, status, size int, duration time.Duration) {
	hlog.FromRequest(r).Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status", status).
		Str("host", r.Host).
		Int("size", size).
		Dur("duration", duration).
		Str("ip", proxyAwareRemote(r)).
		Str("referer", r.Header.Get("Referer")).
		Str("user_agent", r.Header.Get("User-Agent")).
		Msg("Request Completed")
}

// proxyAwareRemote return the most likely remote address
func proxyAwareRemote(r *http.Request) string {
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
