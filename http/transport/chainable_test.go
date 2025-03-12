// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestRoundTripperRace will detect race conditions
// in any RoundTripper by sending concurrent requests.
// Make sure to use the -race parameter when
// executing this test.
func TestRoundTripperRace(t *testing.T) {
	client := http.Client{
		Transport: NewDefaultTransportChain(),
	}

	slowOKHandler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}

	router := mux.NewRouter()
	router.HandleFunc("/test001", slowOKHandler)
	router.HandleFunc("/test002", slowOKHandler)

	server := httptest.NewServer(router)

	go func() {
		for range 10 {
			resp, err := client.Get(server.URL + "/test001")
			if err == nil {
				_ = resp.Body.Close()
			}
		}
	}()

	for range 10 {
		resp, err := client.Get(server.URL + "/test002")
		if err == nil {
			_ = resp.Body.Close()
		}
	}
}

func TestRoundTripperChaining(t *testing.T) {
	t.Run("Chain empty", func(t *testing.T) {
		transport := &recordingTransport{}
		c := Chain().Final(transport)

		url := "/foo"
		req := httptest.NewRequest(http.MethodGet, url, nil)

		_, err := c.RoundTrip(req) //nolint:bodyclose
		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if v := transport.req.Method; v != http.MethodGet {
			t.Errorf("Expected method %q, got %q", http.MethodGet, v)
		}

		if v := transport.req.URL.String(); v != url {
			t.Errorf("Expected URL %q, got %q", url, v)
		}
	})
	t.Run("Chain contains one element", func(t *testing.T) {
		transport := &recordingTransport{}
		c := Chain()
		c.Use(&addHeaderRoundTripper{key: "foo", value: "bar"}).Final(transport)

		url := "/foo"
		req := httptest.NewRequest(http.MethodGet, url, nil)

		_, err := c.RoundTrip(req) //nolint:bodyclose
		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if v := transport.req.Method; v != http.MethodGet {
			t.Errorf("Expected method %v, got %v", http.MethodGet, v)
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
		c := Chain(rt1, rt2, rt3).Final(transport)

		url := "/foo"
		req := httptest.NewRequest(http.MethodGet, url, nil)

		_, err := c.RoundTrip(req) //nolint:bodyclose
		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		if v := transport.req.Method; v != http.MethodGet {
			t.Errorf("Expected method %v, got %v", http.MethodGet, v)
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
