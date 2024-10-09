// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	_ "github.com/pace/bricks/maintenance/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingRoundTripper(t *testing.T) {
	err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	require.NoError(t, err)

	t.Run("With successful response", func(t *testing.T) {
		l := &TracingRoundTripper{}
		tr := &recordingTransportWithResponse{statusCode: 202}
		l.SetTransport(tr)

		req := httptest.NewRequest("GET", "/foo", nil)

		_, err := l.RoundTrip(req)
		require.NoError(t, err)

		require.NotNil(t, tr.span)

		_, ok := tr.span.Data["attempt"]
		assert.False(t, ok)

		code, ok := tr.span.Data["code"]
		assert.True(t, ok)
		assert.Equal(t, 202, code)

		assert.Equal(t, "GET /foo", tr.span.Name)
	})

	t.Run("With error response", func(t *testing.T) {
		l := &TracingRoundTripper{}
		e := errors.New("some error")
		tr := &recordingTransportWithError{err: e}
		l.SetTransport(tr)

		req := httptest.NewRequest("GET", "/bar", nil)
		_, err := l.RoundTrip(req)

		assert.Equal(t, err, e)
		assert.Equal(t, "GET /bar", tr.span.Name)

		val, ok := tr.span.Data["error"]
		assert.True(t, ok)
		assert.Equal(t, val, e)
	})

	t.Run("With retries", func(t *testing.T) {
		tr := &retriedTransport{statusCodes: []int{502, 503, 200}}
		l := Chain(NewDefaultRetryRoundTripper(), &TracingRoundTripper{})
		l.Final(tr)

		req := httptest.NewRequest("GET", "/bar", nil)

		_, err := l.RoundTrip(req)
		require.NoError(t, err)

		span := sentry.TransactionFromContext(tr.ctx)
		require.NotNil(t, span)

		assert.Equal(t, "GET /bar", span.Name)

		val, ok := span.Data["attempt"]
		assert.True(t, ok)
		assert.Equal(t, 3, val)
	})
}

type recordingTransportWithResponse struct {
	span       *sentry.Span
	statusCode int
}

func (t *recordingTransportWithResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	t.span = sentry.TransactionFromContext(req.Context())
	resp := &http.Response{StatusCode: t.statusCode}

	return resp, nil
}

type recordingTransportWithError struct {
	span *sentry.Span
	err  error
}

func (t *recordingTransportWithError) RoundTrip(req *http.Request) (*http.Response, error) {
	t.span = sentry.TransactionFromContext(req.Context())

	return nil, t.err
}
