// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/15 by Florian Hübsch

package transport

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetryRoundTripper(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo", nil)

	t.Run("Successful response after some retries", func(t *testing.T) {
		rt := NewDefaultRetryRoundTripper()
		tr := &retriedTransport{body: "abc", statusCodes: []int{408, 502, 503, 504, 200}}
		rt.SetTransport(tr)

		resp, err := rt.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if ex, got := 4, tr.attempts; got != ex {
			t.Errorf("Expected %d attempts, got %d", ex, got)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Expected readable body, got error: %q", err.Error())
		}
		if tr.body != string(body) {
			t.Errorf("Expected body %q, got %q", tr.body, string(body))
		}
		if got, ex := attemptFromCtx(tr.ctx), int32(4); got != ex {
			t.Errorf("Expected %d attempts, got %d", ex, got)
		}
	})
	t.Run("No retry after error response", func(t *testing.T) {
		rt := NewDefaultRetryRoundTripper()
		e := errors.New("abc")
		tr := &retriedTransport{err: e}
		rt.SetTransport(tr)

		_, err := rt.RoundTrip(req)

		if err == nil {
			t.Fatal("Expected error to be returned, got nil")
		}
		if got, ex := err.Error(), e.Error(); got != ex {
			t.Errorf("Expected error %q, got %q", ex, got)
		}
		if ex, got := 1, tr.attempts; got != ex {
			t.Errorf("Expected %d attempts, got %d", ex, got)
		}
		if got, ex := attemptFromCtx(tr.ctx), int32(1); got != ex {
			t.Errorf("Expected %d attempts, got %d", ex, got)
		}
	})
}

type retriedTransport struct {
	// number of attempts
	attempts int
	// returned status codes in order they are provided
	statusCodes []int
	// returned response body as string
	body string
	// returned error
	err error
	// recorded context
	ctx context.Context
}

func (t *retriedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.attempts++
	t.ctx = req.Context()

	if t.err != nil {
		return nil, t.err
	}
	body := ioutil.NopCloser(bytes.NewReader([]byte(t.body)))
	resp := &http.Response{Body: body, StatusCode: t.statusCodes[t.attempts]}

	return resp, nil
}
