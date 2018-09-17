// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

package oauth2

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Example() {
	r := mux.NewRouter()
	middleware := Middleware{
		URL:          "http://localhost:3000",
		ClientID:     "13972c02189a6e938a4730bc81c2a20cc4e03ef5406d20d2150110584d6b3e6c",
		ClientSecret: "7d26f8918a83bd155a936bbe780f32503a88cb8bd3e8acf25248357dff31668e",
	}

	r.Use(middleware.Handler)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userid, _ := UserID(r.Context())
		log.Printf("AUDIT: User %s does something", userid)

		if HasScope(r.Context(), "dtc:codes:write") {
			fmt.Fprintf(w, "User has scope.")
			return
		}

		fmt.Fprintf(w, "Your client may not have the right scopes to see the secret code")
	})

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
	}

	log.Fatal(srv.ListenAndServe())
}

func TestRequest(t *testing.T) {
	var to = token{
		value:    "somevalue",
		userID:   "someuserid",
		clientID: "someclientid",
		scopes:   []string{"scope1 scope2"},
	}

	r := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := context.WithValue(r.Context(), tokenKey, &to)
	r = r.WithContext(ctx)

	r2 := Request(r)
	header := r2.Header.Get("Authorization")

	if header != "Bearer somevalue" {
		t.Fatalf("Expected request to have authorization header, got: %v", header)
	}
}

func TestRequestWithNoToken(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com", nil)
	r2 := Request(r)
	header := r2.Header.Get("Authorization")

	if header != "" {
		t.Fatalf("Expected request to have no authorization header, got: %v", header)
	}
}
