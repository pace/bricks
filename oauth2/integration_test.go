// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

package oauth2

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	oauthURL      = "https://cp-1-dev.pacelink.net"
	oauthClient   = "7d51282118633c3a7412d7456368ddfe172b7987d20b8e3e60ae18e8681fac61"
	oauthSecret   = "141f891391d2b529bbf37b5ae5f57000f8b093956121db51c90fefb83930175c"
	activeToken   = "85b7856f3055411c11b60f582fc091a624db4a38218abac2df9feb66bc6e7eb5"
	userID        = "b773de39-93d8-4aa4-94a3-356900e55956"
	allowedScopes = "dtc:codes:read dtc:codes:write"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func introspectMock(m *Middleware, token string, s *introspectResponse) error {
	s.UserID = userID
	s.Scope = allowedScopes
	s.ClientID = oauthClient
	s.Active = true
	return nil
}

func TestIntegrationMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test due to -short")
		return
	}

	var middleware = Middleware{
		URL:          oauthURL,
		ClientID:     oauthClient,
		ClientSecret: oauthSecret,
	}

	middleware.addIntrospectFunc(introspectMock)

	router := mux.NewRouter()
	router.Use(middleware.Handler)
	router.HandleFunc("/broken", dummyHandler)
	router.HandleFunc("/inactive", dummyHandler)

	// Test no token.
	rw := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/broken", nil)
	router.ServeHTTP(rw, req)

	if rw.Body.String() != "Unauthorized\n" {
		t.Fatalf("Expected `Unauthorized` as body when *no* token is sent, got %s.", rw.Body)
	}

	// Check for data we are interested in inside the context.
	testMiddlewareHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if we have the X-UID.
		uid, _ := UserID(r.Context())
		if rw.Result().StatusCode != 200 || uid != userID {
			t.Fatal("Expected successful request and X-UID stored in request context.")
		}

		// Check if we have the token.
		receivedToken, _ := BearerToken(r.Context())

		if receivedToken != activeToken {
			t.Fatalf("Expected %s, got: %s", activeToken, receivedToken)
		}

		// Check if we have the scopes.
		scopes := Scopes(r.Context())

		if len(scopes) < 2 {
			t.Fatal("Expected scopes: dtc:codes:read and dtc:codes:write, got nothing.")
		}

		if scopes[0] != "dtc:codes:read" || scopes[1] != "dtc:codes:write" {
			t.Fatalf("Expected scopes: dtc:codes:read and dtc:codes:write, got: %s", scopes)
		}

		// Check if we have the client ID.
		clientID, _ := ClientID(r.Context())

		if clientID != oauthClient {
			t.Fatalf("Expected ClientID %s, got: %s", oauthClient, clientID)
		}
	}

	rw = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/working", nil)
	req.Header.Set("Authorization", "Bearer "+activeToken)
	router.HandleFunc("/working", testMiddlewareHandler)
	router.ServeHTTP(rw, req)

	// This is a last check to make sure everything is good. We must do this check,
	// because it indirectly ensures that the testMiddlewareHandler did actually
	// run. We do not have other options because our /introspect endpoint does not
	// differentiate between bad and old tokens.
	if rw.Result().StatusCode != 200 || rw.Body.String() == "Unauthorized\n" {
		t.Fatalf("Unexpected results using token: %s, perhaps it expired?", activeToken)
	}
}
