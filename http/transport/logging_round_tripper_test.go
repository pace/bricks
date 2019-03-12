// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/12 by Florian Hübsch

package transport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	zlog "github.com/rs/zerolog"
)

func TestLoggingRoundTripper(t *testing.T) {
	l := &LoggingRoundTripper{}
	l.SetTransport(&transportWithResponse{statusCode: 200})

	// create context with logger, capture log output with `out`
	out := &bytes.Buffer{}
	ctx := zlog.New(out).WithContext(context.Background())

	// create request with context and url
	req := httptest.NewRequest("GET", "/foo", nil).WithContext(ctx)
	url, err := url.Parse("http://example.com/foo")
	if err != nil {
		panic(err)
	}
	req.URL = url

	// execute request
	_, err = l.RoundTrip(req)

	if err != nil {
		t.Fatalf("Expected err to be nil, got %#v", err)
	}

	got := string(out.Bytes())
	if !strings.Contains(got, "duration") {
		t.Errorf("Expected duration to be contained in log output, got %v", got)
	}

	gotWithoutDuration := got[:strings.Index(got, "duration")] + got[strings.Index(got, "code"):]
	ex := `{"level":"debug","url":"http://example.com/foo","method":"GET","code":200,"message":"HTTP GET example.com"}`
	if !strings.Contains(gotWithoutDuration, ex) {
		t.Errorf("Expected %v to be contained in log output, got %v", ex, gotWithoutDuration)
	}
}

type transportWithResponse struct {
	statusCode int
}

func (t *transportWithResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: t.statusCode}

	return resp, nil
}
