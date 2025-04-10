package transport

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreakerTripper(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	t.Run("with_default_settings", func(t *testing.T) {
		breaker := NewDefaultCircuitBreakerTripper("testcircuitbreaker")
		chain := Chain(breaker).Final(&failingRoundTripper{})

		for range 6 {
			_, err := chain.RoundTrip(req) //nolint:bodyclose
			require.NotErrorIs(t, err, ErrCircuitBroken)
		}

		_, err := chain.RoundTrip(req) //nolint:bodyclose
		require.ErrorIs(t, err, ErrCircuitBroken)
	})

	t.Run("panic_on_empty_name", func(t *testing.T) {
		assert.Panics(t, assert.PanicTestFunc(func() {
			NewCircuitBreakerTripper(gobreaker.Settings{})
		}), "NewCircuitBreakerTripper did not panic on empty name")

		assert.Panics(t, assert.PanicTestFunc(func() {
			NewDefaultCircuitBreakerTripper("")
		}), "NewDefaultCircuitBreakerTripper did not panic on empty name")
	})

	t.Run("resp_object_untouched", func(t *testing.T) {
		wantBodyStr := "some text in the body which should be unaffected by the circuit breaker"
		breaker := NewDefaultCircuitBreakerTripper("testcircuitbreaker")
		chain := Chain(breaker).Final(&transportWithBody{body: wantBodyStr})

		resp, err := chain.RoundTrip(req)
		require.NoError(t, err, "expected no err, got err=%q", err)

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

		gotBodyStr, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "failed reading response body no err, got err=%q", err)

		if string(gotBodyStr) != wantBodyStr {
			t.Errorf("request and response body do not match, wanted=%q, got=%q", wantBodyStr, string(gotBodyStr))
		}
	})
}

type failingRoundTripper struct{}

func (f *failingRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("connection error")
}
