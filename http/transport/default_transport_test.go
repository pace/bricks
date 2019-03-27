// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/26 by Florian Hübsch

package transport

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDefaultTransport(t *testing.T) {
	t.Run("Finalizer nil", func(t *testing.T) {
		b := "Hello World"
		tr := NewDefaultTransport(nil)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, b)
		}))
		defer ts.Close()

		req := httptest.NewRequest("GET", ts.URL, nil)
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
		tr := &transportWithBody{body: "abc"}
		dt := NewDefaultTransport(tr)

		req := httptest.NewRequest("GET", "/foo", nil)
		resp, err := dt.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Expected readable body, got error: %q", err.Error())
		}
		if ex, got := tr.body, string(body); ex != got {
			t.Errorf("Expected body %q, got %q", ex, got)
		}
	})
}

type transportWithBody struct {
	// returned response as string
	body string
}

func (t *transportWithBody) RoundTrip(req *http.Request) (*http.Response, error) {
	body := ioutil.NopCloser(bytes.NewReader([]byte(t.body)))
	resp := &http.Response{Body: body, StatusCode: 200}

	return resp, nil
}
