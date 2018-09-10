package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"oauth2"
	"testing"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func TestMiddleware(t *testing.T) {
	// TODO
	// Run against cp-1-dev or cp-1-prod?
	middleware := oauth2.Middleware{
		URL:          "http://localhost:3000",
		ClientID:     "13972c02189a6e938a4730bc81c2a20cc4e03ef5406d20d2150110584d6b3e6c",
		ClientSecret: "7d26f8918a83bd155a936bbe780f32503a88cb8bd3e8acf25248357dff31668e",
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
		t.Fatalf("Expected `Unauthorized` as response body when *no* token is provided, got %s.", rw.Body)
	}

	// Test bad token.
	rw = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/broken", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	router.ServeHTTP(rw, req)

	if rw.Body.String() != "Unauthorized\n" {
		t.Fatalf("Expected `Unauthorized` as response body when *bad* token is provided, got %s.", rw.Body)
	}

	// Check for data we are interested in inside the context.
	testMiddlewareHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if we have the X-UID.
		if rw.Result().StatusCode != 200 || oauth2.UserID(r.Context()) != "3298b629-0467-400e-b430-5259cc3efddc" {
			t.Fatal("Expected successful request and X-UID stored in request context.")
		}

		// Check if we have the token.
		receivedToken := oauth2.BearerToken(r.Context())
		expectedToken := "a19fb24b914bdd93b1acc3f796dc52bd0ca1e642fa74f2f4657486e24109cb39"

		if receivedToken != expectedToken {
			t.Fatalf("Expected %s, got: %s", expectedToken, receivedToken)
		}

		// Check if we have the scopes.
		scopes := oauth2.Scopes(r.Context())
		if scopes[0] != "dtc:codes:read" || scopes[1] != "dtc:codes:write" {
			t.Fatalf("Expected scopes: dtc:codes:read and dtc:codes:write, got: %s", scopes)
		}

		expectedClientID := "13972c02189a6e938a4730bc81c2a20cc4e03ef5406d20d2150110584d6b3e6c"

		// Check if we have the client ID.
		clientID := oauth2.ClientID(r.Context())

		if clientID != expectedClientID {
			t.Fatalf("Expected ClientID 6, got: %s", clientID)
		}
	}

	rw = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/working", nil)
	req.Header.Set("Authorization", "Bearer a19fb24b914bdd93b1acc3f796dc52bd0ca1e642fa74f2f4657486e24109cb39")
	router.HandleFunc("/working", testMiddlewareHandler)
	router.ServeHTTP(rw, req)
}
