// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"context"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/locale"
)

type roundTripperMock struct {
	r *http.Request
}

func (m *roundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	m.r = req
	return nil, nil
}

func TestLocaleRoundTrip(t *testing.T) {
	mock := new(roundTripperMock)
	lrt := &LocaleRoundTripper{transport: mock}

	l := locale.NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Europe/Paris")
	r, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
	require.NoError(t, err)

	_, err = lrt.RoundTrip(r.WithContext(locale.WithLocale(context.Background(), l))) //nolint:bodyclose
	require.NoError(t, err)

	lctx, ok := locale.FromCtx(mock.r.Context())
	require.True(t, ok)
	assert.Equal(t, lctx.Language(), "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	assert.Equal(t, lctx.Timezone(), "Europe/Paris")

	data, err := httputil.DumpRequest(mock.r, false)
	require.NoError(t, err)
	assert.Equal(t, "GET /test HTTP/1.1\r\nHost: example.com\r\nAccept-Language: fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5\r\nAccept-Timezone: Europe/Paris\r\n\r\n", string(data))
}
