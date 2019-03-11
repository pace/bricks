// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/11 by Florian Hübsch

package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoundTripperChaining(t *testing.T) {
	t.Run("Chain empty", func(t *testing.T) {
		transport := &recordingTransport{}
		c := &RoundTripperChain{RoundTrippers: []ChainableRoundTripper{}, Transport: transport}

		url := "/foo"
		req := httptest.NewRequest("GET", url, nil)

		_, err := c.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if v := transport.req.Method; v != "GET" {
			t.Errorf("Expected method %q, got %q", "GET", v)
		}
		if v := transport.req.URL.String(); v != url {
			t.Errorf("Expected URL %q, got %q", url, v)
		}
	})
	t.Run("Chain contains one element", func(t *testing.T) {
		transport := &recordingTransport{}
		rt := &addHeaderRoundTripper{key: "foo", value: "bar"}
		c := &RoundTripperChain{RoundTrippers: []ChainableRoundTripper{rt}, Transport: transport}

		url := "/foo"
		req := httptest.NewRequest("GET", url, nil)

		_, err := c.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if v := transport.req.Method; v != "GET" {
			t.Errorf("Expected method %v, got %v", "GET", v)
		}
		if v := transport.req.URL.String(); v != url {
			t.Errorf("Expected URL %v, got %v", url, v)
		}
		if v, ex := transport.req.Header.Get("foo"), "bar"; v != ex {
			t.Errorf("Expected header foo to eq %v, got %v", ex, v)
		}
	})
	t.Run("Chain contains multiple elements", func(t *testing.T) {
		transport := &recordingTransport{}
		rt1 := &addHeaderRoundTripper{key: "foo", value: "bar"}
		rt2 := &addHeaderRoundTripper{key: "foo", value: "baroverride"}
		rt3 := &addHeaderRoundTripper{key: "Authorization", value: "Bearer 123"}
		c := &RoundTripperChain{RoundTrippers: []ChainableRoundTripper{rt1, rt2, rt3}, Transport: transport}

		url := "/foo"
		req := httptest.NewRequest("GET", url, nil)

		_, err := c.RoundTrip(req)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if v := transport.req.Method; v != "GET" {
			t.Errorf("Expected method %v, got %v", "GET", v)
		}
		if v := transport.req.URL.String(); v != url {
			t.Errorf("Expected URL %v, got %v", url, v)
		}
		if v, ex := transport.req.Header.Get("foo"), "baroverride"; v != ex {
			t.Errorf("Expected header foo to eq %v, got %v", ex, v)
		}
		if v, ex := transport.req.Header.Get("Authorization"), "Bearer 123"; v != ex {
			t.Errorf("Expected header Authorization to eq %v, got %v", ex, v)
		}
	})
}

type recordingTransport struct {
	req *http.Request
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req

	return nil, nil
}

type addHeaderRoundTripper struct {
	key       string
	value     string
	transport http.RoundTripper
}

func (r *addHeaderRoundTripper) Transport() http.RoundTripper {
	return r.transport
}

func (r *addHeaderRoundTripper) SetTransport(rt http.RoundTripper) {
	r.transport = rt
}

func (r *addHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(r.key, r.value)

	return r.Transport().RoundTrip(req)
}
