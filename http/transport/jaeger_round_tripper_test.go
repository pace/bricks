// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/14 by Florian Hübsch

package transport

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	_ "github.com/pace/bricks/maintenance/tracing"
)

func TestJaegerRoundTripper(t *testing.T) {
	t.Run("With successful response", func(t *testing.T) {
		l := &JaegerRoundTripper{}
		tr := &recordingTransportWithResponse{statusCode: 202}
		l.SetTransport(tr)

		req := httptest.NewRequest("GET", "/foo", nil)
		_, err := l.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		spanString := fmt.Sprintf("%#v", tr.span)

		if strings.Contains(spanString, "attempt") {
			t.Errorf("Expected attempt to not be included in span %q", spanString)
		}
		exs := []string{`operationName:"GET /foo"`, "numericVal:202"}
		for _, ex := range exs {
			if !strings.Contains(spanString, ex) {
				t.Errorf("Expected %q to be included in span %v", ex, spanString)
			}
		}
	})
	t.Run("With error response", func(t *testing.T) {
		l := &JaegerRoundTripper{}
		e := errors.New("some error")
		tr := &recordingTransportWithError{err: e}
		l.SetTransport(tr)

		req := httptest.NewRequest("GET", "/bar", nil)
		_, err := l.RoundTrip(req)

		if got, ex := err.Error(), e.Error(); got != ex {
			t.Fatalf("Expected error %q to be returned, got %q", ex, got)
		}

		spanString := fmt.Sprintf("%#v", tr.span)
		exs := []string{`operationName:"GET /bar"`, `log.Field{key:"error"`}
		for _, ex := range exs {
			if !strings.Contains(spanString, ex) {
				t.Errorf("Expected %q to be included in span %v", ex, spanString)
			}
		}
	})
	t.Run("With retries", func(t *testing.T) {
		tr := &retriedTransport{statusCodes: []int{502, 503, 200}}
		l := Chain(NewDefaultRetryRoundTripper(), &JaegerRoundTripper{})
		l.Final(tr)

		req := httptest.NewRequest("GET", "/bar", nil)
		_, err := l.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		span := opentracing.SpanFromContext(tr.ctx)
		spanString := fmt.Sprintf("%#v", span)
		exs := []string{`operationName:"GET /bar"`, `log.Field{key:"attempt", fieldType:2, numericVal:3`}
		for _, ex := range exs {
			if !strings.Contains(spanString, ex) {
				t.Errorf("Expected %q to be included in span %v", ex, spanString)
			}
		}
	})
}

type recordingTransportWithResponse struct {
	span       opentracing.Span
	statusCode int
}

func (t *recordingTransportWithResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	t.span = opentracing.SpanFromContext(req.Context())
	resp := &http.Response{StatusCode: t.statusCode}

	return resp, nil
}

type recordingTransportWithError struct {
	span opentracing.Span
	err  error
}

func (t *recordingTransportWithError) RoundTrip(req *http.Request) (*http.Response, error) {
	t.span = opentracing.SpanFromContext(req.Context())

	return nil, t.err
}
