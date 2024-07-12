// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	. "github.com/pace/bricks/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextTransfer(t *testing.T) {
	r, err := http.NewRequest("GET", "http://example.com/", nil)
	require.NoError(t, err)
	r.Header.Set("User-Agent", "Foobar")
	RequestInContext(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctx := ContextTransfer(r.Context(), context.Background())
		userAgent, err := GetUserAgentFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Foobar", userAgent)
	})).ServeHTTP(nil, r)

	// without request
	ctx := ContextTransfer(context.Background(), context.Background())
	userAgent, err := GetUserAgentFromContext(ctx)
	assert.True(t, errors.Is(err, ErrNotFound), err)
	assert.Empty(t, userAgent)
}

func TestGetXForwardedForHeaderFromContext(t *testing.T) {
	cases := map[string]struct {
		RemoteAddr          string // input IP:port
		XForwardedFor       string // input X-Forwarded-For header
		ExpectErr           error
		ExpectXForwardedFor string // output X-Forwarded-For header
	}{
		"direct request": {
			RemoteAddr:          "12.34.56.78:9999",
			ExpectXForwardedFor: "12.34.56.78",
		},
		"behind a proxy": {
			RemoteAddr:          "12.34.56.78:9999",
			XForwardedFor:       "100.100.100.100",
			ExpectXForwardedFor: "100.100.100.100, 12.34.56.78",
		},
		"behind multiple proxies": {
			RemoteAddr:          "4.4.4.4:1234",
			XForwardedFor:       "1.1.1.1, 2.2.2.2, 3.3.3.3",
			ExpectXForwardedFor: "1.1.1.1, 2.2.2.2, 3.3.3.3, 4.4.4.4",
		},
		"ipv6": {
			RemoteAddr:          "[d953:7242:7970:566c:ee3a:0581:36cd:4fd6]:1234",
			XForwardedFor:       "7342:57fb:4188:fd49:1eed:644f:22d6:69a2",
			ExpectXForwardedFor: "7342:57fb:4188:fd49:1eed:644f:22d6:69a2, d953:7242:7970:566c:ee3a:0581:36cd:4fd6",
		},
		"missing remote address": {
			RemoteAddr: "",
			ExpectErr:  ErrInvalidRequest,
		},
		"missing remote ip": {
			RemoteAddr: ":80",
			ExpectErr:  ErrInvalidRequest,
		},
		"broken remote address": {
			RemoteAddr: "1234567890ß",
			ExpectErr:  ErrInvalidRequest,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "http://example.com/", nil)
			require.NoError(t, err)
			r.RemoteAddr = c.RemoteAddr
			if c.XForwardedFor != "" {
				r.Header.Set("X-Forwarded-For", c.XForwardedFor)
			}
			RequestInContext(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				xForwardedFor, err := GetXForwardedForHeaderFromContext(ctx)
				if c.ExpectErr != nil {
					assert.True(t, errors.Is(err, c.ExpectErr))
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, c.ExpectXForwardedFor, xForwardedFor)
			})).ServeHTTP(nil, r)
		})
	}

	// no request in context
	xForwardedFor, err := GetXForwardedForHeaderFromContext(context.Background())
	assert.True(t, errors.Is(err, ErrNotFound), err)
	assert.Empty(t, xForwardedFor)
}

func TestGetUserAgentFromContext(t *testing.T) {
	r, err := http.NewRequest("GET", "http://example.com/", nil)
	require.NoError(t, err)
	r.Header.Set("User-Agent", "Foobar")
	RequestInContext(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userAgent, err := GetUserAgentFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Foobar", userAgent)
	})).ServeHTTP(nil, r)

	// no request in context
	userAgent, err := GetUserAgentFromContext(context.Background())
	assert.True(t, errors.Is(err, ErrNotFound), err)
	assert.Empty(t, userAgent)
}
