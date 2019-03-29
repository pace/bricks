// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/28 by Florian Hübsch

package middleware

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mux "github.com/gorilla/mux"
	"github.com/pace/bricks/http/oauth2"
)

func TestScopesMiddleware(t *testing.T) {
	t.Run("Token scope sufficient for endpoint", func(t *testing.T) {
		r := setupRouter("foo:read", "foo:write foo:read")
		req := setupRequest()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		if got, ex := resp.StatusCode, 200; got != ex {
			t.Errorf("Expected status code %d, got %d", ex, got)
		}

		if got, ex := string(body), "Hello"; got != ex {
			t.Errorf("Expected body %q, got %q", ex, got)
		}
	})

	t.Run("Token scope insufficient for endpoint", func(t *testing.T) {
		r := setupRouter("foo:read", "foo:write")
		req := setupRequest()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		if got, ex := resp.StatusCode, 401; got != ex {
			t.Errorf("Expected status code %d, got %d", ex, got)
		}

		if got, ex := string(body), fmt.Sprintf("Unauthorized - requires scope %q\n", "foo:read"); got != ex {
			t.Errorf("Expected body %q, got %q", ex, got)
		}
	})
}

func setupRouter(requiredScope string, tokenScope string) *mux.Router {
	rs := RequiredScopes{
		"GetFoo": oauth2.Scope(requiredScope),
	}
	m := NewScopesMiddleware(rs)
	om := oauth2.NewMiddleware(&tokenIntrospecter{returnedScope: tokenScope})

	r := mux.NewRouter()
	r.Use(om.Handler)
	r.Use(m.Handler)
	r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello")
	}).Name("GetFoo")

	return r
}

func setupRequest() *http.Request {
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Authorization", "Bearer some-token")

	return req
}

type tokenIntrospecter struct {
	returnedScope string
}

func (t *tokenIntrospecter) IntrospectToken(ctx context.Context, token string) (*oauth2.IntrospectResponse, error) {
	resp := &oauth2.IntrospectResponse{Active: true, Scope: t.returnedScope}
	return resp, nil
}
