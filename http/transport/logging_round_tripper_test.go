// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pace/bricks/maintenance/log"
)

func TestLoggingRoundTripper(t *testing.T) {
	// create context with logger, capture log output with `out`
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	// create request with context and url
	req := httptest.NewRequest(http.MethodGet, "/foo", nil).WithContext(ctx)

	url, err := url.Parse("http://example.com/foo")
	if err != nil {
		panic(err)
	}

	req.URL = url

	t.Run("Without retries", func(t *testing.T) {
		l := &LoggingRoundTripper{}
		l.SetTransport(&transportWithResponse{statusCode: http.StatusOK})

		resp, err := l.RoundTrip(req)
		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

		got := out.String()
		if !strings.Contains(got, "duration") {
			t.Errorf("Expected duration to be contained in log output, got %v", got)
		}

		if strings.Contains(got, "retries") {
			t.Errorf("Expected retries to not be contained in log output, got %v", got)
		}

		exs := []string{`"level":"debug"`, `"url":"http://example.com/foo"`, `"method":"GET"`, `"status_code":200`, `"message":"HTTP GET example.com"`}
		for _, ex := range exs {
			if !strings.Contains(got, ex) {
				t.Errorf("Expected %v to be contained in log output, got %v", ex, got)
			}
		}
	})
	t.Run("With retries", func(t *testing.T) {
		l := Chain(NewDefaultRetryRoundTripper(), &LoggingRoundTripper{})
		l.Final(&retriedTransport{statusCodes: []int{502, 503, 408, 202}})

		resp, err := l.RoundTrip(req)
		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

		got := out.String()
		exs := []string{`"level":"debug"`, `"url":"http://example.com/foo"`, `"method":"GET"`, `"status_code":200`, `"message":"HTTP GET example.com"`, `"attempt":3`}

		for _, ex := range exs {
			if !strings.Contains(got, ex) {
				t.Errorf("Expected %v to be contained in log output, got %v", ex, got)
			}
		}
	})
}
