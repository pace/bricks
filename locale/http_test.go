// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package locale

import (
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)

	l := FromRequest(r)
	assert.False(t, l.HasLanguage())
	assert.False(t, l.HasTimezone())
}

func TestFilledRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)
	r.Header.Set(HeaderAcceptLanguage, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	r.Header.Set(HeaderAcceptTimezone, "Europe/Paris")

	l := FromRequest(r)
	assert.Equal(t, l.Language(), "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	assert.Equal(t, l.Timezone(), "Europe/Paris")
}

func TestExtendRequestWithEmptyLocale(t *testing.T) {
	l := new(Locale)
	r, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)

	data, err := httputil.DumpRequest(l.Request(r), false)
	require.NoError(t, err)

	assert.Equal(t, "GET /test HTTP/1.1\r\nHost: example.com\r\n\r\n", string(data))
}

func TestExtendRequest(t *testing.T) {
	l := NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Europe/Paris")
	r, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)

	data, err := httputil.DumpRequest(l.Request(r), false)
	require.NoError(t, err)

	assert.Equal(t, "GET /test HTTP/1.1\r\nHost: example.com\r\nAccept-Language: fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5\r\nAccept-Timezone: Europe/Paris\r\n\r\n", string(data))
}

type httpRecorderNext struct {
	w http.ResponseWriter
	r *http.Request
}

func (m *httpRecorderNext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.w = w
	m.r = r
}

func TestMiddlewareWithoutLocale(t *testing.T) {
	r, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)

	rec := new(httpRecorderNext)
	h := Handler()(rec)
	h.ServeHTTP(nil, r)

	lctx, ok := FromCtx(rec.r.Context())
	require.True(t, ok)
	assert.False(t, lctx.HasLanguage())
	assert.False(t, lctx.HasTimezone())
}

func TestMiddlewareWithLocale(t *testing.T) {
	l := NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Europe/Paris")
	r, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)

	rec := new(httpRecorderNext)
	h := Handler()(rec)
	h.ServeHTTP(nil, l.Request(r))

	lctx, ok := FromCtx(rec.r.Context())
	require.True(t, ok)
	assert.Equal(t, lctx.Language(), "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	assert.Equal(t, lctx.Timezone(), "Europe/Paris")
}
