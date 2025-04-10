// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package log

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if RequestID(r) == "" {
			t.Error("Request should have request id")
		}

		w.WriteHeader(http.StatusCreated)
	})
	Handler()(mux).ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		t.Error("expected 201 status code")
	}
}

func Test_isPrivate(t *testing.T) {
	tests := []struct {
		ip   net.IP
		want bool
	}{
		{nil, false},
		{net.IPv4(10, 0, 0, 0), true},
		{net.IPv4(11, 0, 0, 0), false},
		{net.IPv4(172, 16, 0, 0), true},
		{net.IPv4(172, 32, 0, 0), false},
		{net.IPv4(192, 168, 0, 0), true},
		{net.IPv4(192, 169, 0, 0), false},
		{net.IP{0xfc, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, true},
		{net.IP{0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.ip.String(), func(t *testing.T) {
			if got := isPrivate(tt.ip); got != tt.want {
				t.Errorf("isPrivate() = %v, want %v", got, tt.want)
			}
		})
	}
}
