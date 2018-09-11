package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"oauth2"
	"testing"
)

// Constants used throughout tests.
const (
	oauthURL    = "http://localhost:3000"
	oauthClient = "13972c02189a6e938a4730bc81c2a20cc4e03ef5406d20d2150110584d6b3e6c"
	oauthSecret = "7d26f8918a83bd155a936bbe780f32503a88cb8bd3e8acf25248357dff31668e"
	activeToken = "c58b66b2a1b9b79376b587d68e1090e0d976d2013786ec2f1f49116eab4d62a7"
	userID      = "3298b629-0467-400e-b430-5259cc3efddc"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func TestMiddleware(t *testing.T) {
	// TODO
	// Run against cp-1-dev or cp-1-prod?
	var middleware = oauth2.Middleware{
		URL:          oauthURL,
		ClientID:     oauthClient,
		ClientSecret: oauthSecret,
	}

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

	// Test bad token.
	rw = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/broken", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	router.ServeHTTP(rw, req)

	if rw.Body.String() != "Unauthorized\n" {
		t.Fatalf("Expected `Unauthorized` as body when *bad* token is sent, got %s.", rw.Body)
	}

	// Check for data we are interested in inside the context.
	testMiddlewareHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if we have the X-UID.
		if rw.Result().StatusCode != 200 || oauth2.UserID(r.Context()) != userID {
			t.Fatal("Expected successful request and X-UID stored in request context.")
		}

		// Check if we have the token.
		receivedToken := oauth2.BearerToken(r.Context())

		if receivedToken != activeToken {
			t.Fatalf("Expected %s, got: %s", activeToken, receivedToken)
		}

		// Check if we have the scopes.
		scopes := oauth2.Scopes(r.Context())
		if scopes[0] != "dtc:codes:read" || scopes[1] != "dtc:codes:write" {
			t.Fatalf("Expected scopes: dtc:codes:read and dtc:codes:write, got: %s", scopes)
		}

		// Check if we have the client ID.
		clientID := oauth2.ClientID(r.Context())

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
		t.Fatalf("Unexpected results using token: %s, perhaps it expired?", "token")
	}
}
