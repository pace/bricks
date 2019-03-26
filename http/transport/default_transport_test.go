// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/26 by Florian Hübsch

package transport

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/pace/bricks/maintenance/log"
)

func TestNewDefaultTransport(t *testing.T) {
	// create context with logger, capture log output with `out`
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	t.Run("Finalizer nil", func(t *testing.T) {
		b := "Hello World"
		tr := NewDefaultTransport(nil)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, b)
		}))
		defer ts.Close()

		req := httptest.NewRequest("GET", ts.URL, nil).WithContext(ctx)
		resp, err := tr.RoundTrip(req)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if ex, got := b, string(body); ex != got {
			t.Errorf("Expected body %q, got %q", ex, got)
		}
	})

	t.Run("Finalizer given", func(t *testing.T) {
		tr := &retriedTransport{body: "abc", statusCodes: []int{408, 502, 503, 504, 200}}
		dt := NewDefaultTransport(tr)

		// create request with context and url
		req := httptest.NewRequest("GET", "/foo", nil).WithContext(ctx)
		url, err := url.Parse("http://example.com/foo")
		if err != nil {
			panic(err)
		}
		req.URL = url
		resp, err := dt.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Expected readable body, got error: %q", err.Error())
		}
		if tr.body != string(body) {
			t.Errorf("Expected body %q, got %q", tr.body, string(body))
		}

		// test retry attempts
		if got, ex := attemptFromCtx(tr.ctx), int32(4); got != ex {
			t.Errorf("Expected %d attempts, got %d", ex, got)
		}

		// test jaeger
		span := opentracing.SpanFromContext(tr.ctx)
		spanString := fmt.Sprintf("%#v", span)
		exs := []string{`operationName:"GET /foo"`, `log.Field{key:"attempt", fieldType:2, numericVal:4`, "numericVal:200"}
		for _, ex := range exs {
			if !strings.Contains(spanString, ex) {
				t.Errorf("Expected %q to be included in span %v", ex, spanString)
			}
		}

		// test logging
		got := out.String()
		if !strings.Contains(got, "duration") {
			t.Errorf("Expected duration to be contained in log output, got %v", got)
		}
		if strings.Contains(got, "retries") {
			t.Errorf("Expected retries to not be contained in log output, got %v", got)
		}

		exs = []string{`"level":"debug"`, `"url":"http://example.com/foo"`, `"method":"GET"`, `"code":200`, `"message":"HTTP GET example.com"`}
		for _, ex := range exs {
			if !strings.Contains(got, ex) {
				t.Errorf("Expected %v to be contained in log output, got %v", ex, got)
			}
		}
	})
}
