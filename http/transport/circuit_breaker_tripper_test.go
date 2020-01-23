package transport

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreakerTripper(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo", nil)

	t.Run("with_default_settings", func(t *testing.T) {
		breaker := NewDefaultCircuitBreakerTripper("testcircuitbreaker")
		chain := Chain(breaker).Final(&failingRoundTripper{})

		for i := 0; i < 6; i++ {
			if _, err := chain.RoundTrip(req); errors.Is(err, gobreaker.ErrOpenState) {
				t.Errorf("got err=%q, before expected", gobreaker.ErrOpenState)
			}
		}

		if _, err := chain.RoundTrip(req); !errors.Is(err, gobreaker.ErrOpenState) {
			t.Errorf("wanted err=%q, got err=%q", gobreaker.ErrOpenState, err)
		}
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
		assert.NoError(t, err, "expected no err, got err=%q", err)

		gotBodyStr, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err, "failed reading response body no err, got err=%q", err)

		if string(gotBodyStr) != wantBodyStr {
			t.Errorf("request and response body do not match, wanted=%q, got=%q", wantBodyStr, string(gotBodyStr))
		}
	})
}

type failingRoundTripper struct{}

func (f *failingRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("connection error")
}
