// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Mohamed Wael Khobalatte

package oauth2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessfulIntrospection(t *testing.T) {
	// Start local  server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		expectedPath := "/oauth2/introspect"
		actualPath := req.URL.String()

		if expectedPath != actualPath {
			t.Fatalf("Expected `%s`, got: %s", expectedPath, actualPath)
		}

		rw.Header().Set("Content-Type", "application/json")

		data := introspectResponse{
			Active:   true,
			Scope:    "dtc:codes:read dtc:codes:write",
			ClientID: "SOME_CLIENT_ID",
		}

		// This is what Cockpit does at the moment.
		rw.Header().Set("X-UID", "SOME_USER_ID")

		// Send response to be tested
		json.NewEncoder(rw).Encode(data) //nolint:errcheck
	}))

	defer server.Close()

	var s introspectResponse
	var m = Middleware{
		URL:          server.URL,
		ClientID:     "oauthClient",
		ClientSecret: "oauthSecret",
	}

	err := introspect(&m, "token", &s)

	if err != nil {
		t.Fatalf("Expected no error, got %v.", err)
	}

	if !s.Active || s.Scope != "dtc:codes:read dtc:codes:write" || s.UserID != "SOME_USER_ID" ||
		s.ClientID != "SOME_CLIENT_ID" {
		t.Fatalf("Expected specific values stored in struct, got: %v", s)
	}
}

func TestNoConnection(t *testing.T) {
	var s introspectResponse
	var m = Middleware{
		URL:          "", // No URL mimicks a server down, to some extent.
		ClientID:     "oauthClient",
		ClientSecret: "oauthSecret",
	}

	err := introspect(&m, "token", &s)

	if err != errUpstreamConnection {
		t.Fatalf("Expected error %v, got %v.", errUpstreamConnection, err)
	}
}

func Test400UpstreamResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "BadRequest", http.StatusBadRequest)
	}))

	defer server.Close()

	var s introspectResponse
	var m = Middleware{
		URL:          server.URL,
		ClientID:     "oauthClient",
		ClientSecret: "oauthSecret",
	}

	err := introspect(&m, "token", &s)

	if err != errBadUpstreamResponse {
		t.Fatalf("Expected error %v, got %v.", errBadUpstreamResponse, err)
	}
}

func TestUnparsableUpstreamResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("Not JSON.")) //nolint:errcheck
	}))

	defer server.Close()

	var s introspectResponse
	var m = Middleware{
		URL:          server.URL,
		ClientID:     "oauthClient",
		ClientSecret: "oauthSecret",
	}

	err := introspect(&m, "token", &s)

	if err != errBadUpstreamResponse {
		t.Fatalf("Expected error %v, got %v.", errBadUpstreamResponse, err)
	}
}

func TestBadTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		data := introspectResponse{
			Active: false,
		}

		// Send response to be tested
		json.NewEncoder(rw).Encode(data) //nolint:errcheck
	}))

	defer server.Close()

	var s introspectResponse
	var m = Middleware{
		URL:          server.URL,
		ClientID:     "oauthClient",
		ClientSecret: "oauthSecret",
	}

	err := introspect(&m, "token", &s)

	if err != errInvalidToken {
		t.Fatalf("Expected error %v, got %v.", errInvalidToken, err)
	}
}
