// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

package oauth2

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/log"
)

type tokenInspectorWithError struct {
	returnedErr error
}

func (t *tokenInspectorWithError) IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error) {
	return nil, t.returnedErr
}

func TestHandlerIntrospectErrorAsMiddleware(t *testing.T) {
	testCases := []struct {
		desc         string
		returnedErr  error
		expectedCode int
		expectedBody string
	}{
		{
			desc:         "token introspecter returns ErrBadUpstreamResponse",
			returnedErr:  ErrBadUpstreamResponse,
			expectedCode: 502,
			expectedBody: "bad upstream response when introspecting token\n",
		},
		{
			desc:         "token introspecter returns ErrUpstreamConnection",
			returnedErr:  ErrUpstreamConnection,
			expectedCode: 502,
			expectedBody: "problem connecting to the introspection endpoint\n",
		},
		{
			desc:         "token introspecter returns ErrInvalidToken",
			returnedErr:  ErrInvalidToken,
			expectedCode: 401,
			expectedBody: "user token is invalid\n",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := NewMiddleware(&tokenInspectorWithError{returnedErr: tC.returnedErr})
			r := mux.NewRouter()
			r.Use(m.Handler)
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer some-token")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			if got, ex := resp.StatusCode, tC.expectedCode; got != ex {
				t.Errorf("Expected status code %d, got %d", ex, got)
			}

			if got, ex := string(body), tC.expectedBody; got != ex {
				t.Errorf("Expected body %q, got %q", ex, got)
			}
		})
	}
}

type tokenIntrospectedSuccessful struct {
	response *IntrospectResponse
}

func (t *tokenIntrospectedSuccessful) IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error) {
	return t.response, nil
}

func TestAuthenticatorWithSuccess(t *testing.T) {
	testCases := []struct {
		desc           string
		userScopes     string
		expectedScopes string
		active         bool
		clientId       string
		userId         string
	}{
		{desc: "Tests a valid Request with OAuth2 Authentication without Scope checking",
			active:     true,
			userScopes: "ABC DHHG kjdk",
			clientId:   "ClientId",
			userId:     "UserId",
		},
		{desc: "Tests a valid Request with OAuth2 Authentication and one scope to check",
			active:         true,
			userScopes:     "ABC DHHG kjdk",
			clientId:       "ClientId",
			userId:         "UserId",
			expectedScopes: "ABC",
		},
		{desc: "Tests a valid Request with OAuth2 Authentication and two scope to check",
			active:         true,
			userScopes:     "ABC DHHG kjdk",
			clientId:       "ClientId",
			userId:         "UserId",
			expectedScopes: "ABC kjdk",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Add("Authorization", "Bearer bearer")

			auth := NewAuthorizer(&tokenIntrospectedSuccessful{&IntrospectResponse{
				Active:   tC.active,
				Scope:    tC.userScopes,
				ClientID: tC.clientId,
				UserID:   tC.userId,
			}}, &Config{})
			if tC.expectedScopes != "" {
				auth = auth.WithScope(tC.expectedScopes)
			}
			authorize, b := auth.Authorize(r, w)
			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if !b || authorize == nil {
				t.Errorf("Expected succesfull Authentication, but was not succesfull with code %d and body %q", resp.StatusCode, string(body))
				return
			}
			to, _ := security.GetTokenFromContext(authorize)
			tok, ok := to.(*token)

			if !ok || tok.value != "bearer" || tok.scope != Scope(tC.userScopes) || tok.clientID != tC.clientId || tok.userID != tC.userId {
				t.Errorf("Expected %v but got %v", auth.introspection.(*tokenIntrospectedSuccessful).response, tok)
			}
		})
	}
}

func TestAuthenticationSuccessScopeError(t *testing.T) {
	auth := NewAuthorizer(&tokenIntrospectedSuccessful{&IntrospectResponse{
		Active:   true,
		Scope:    "ABC DEF DFE",
		ClientID: "ClientId",
		UserID:   "UserId",
	}}, &Config{}).WithScope("DE")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "Bearer bearer")

	_, b := auth.Authorize(r, w)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Errorf("Expected error in Authentication, but was succesfull with code %d and body %v", resp.StatusCode, string(body))
	}
	if got, ex := w.Code, http.StatusForbidden; got != ex {
		t.Errorf("Expected status code %d, got %d", ex, got)
	}
	if got, ex := string(body), "Forbidden - requires scope \"DE\"\n"; got != ex {
		t.Errorf("Expected status code %q, got %q", ex, got)
	}
}

func TestAuthenticationWithErrors(t *testing.T) {
	testCases := []struct {
		desc         string
		returnedErr  error
		expectedCode int
		expectedBody string
	}{
		{
			desc:         "token introspecter returns ErrBadUpstreamResponse",
			returnedErr:  ErrBadUpstreamResponse,
			expectedCode: http.StatusBadGateway,
			expectedBody: "bad upstream response when introspecting token\n",
		},
		{
			desc:         "token introspecter returns ErrUpstreamConnection",
			returnedErr:  ErrUpstreamConnection,
			expectedCode: http.StatusBadGateway,
			expectedBody: "problem connecting to the introspection endpoint\n",
		},
		{
			desc:         "token introspecter returns ErrInvalidToken",
			returnedErr:  ErrInvalidToken,
			expectedCode: http.StatusUnauthorized,
			expectedBody: "user token is invalid\n",
		},
		{
			desc:         "token introspecter returns any other error",
			returnedErr:  errors.New("any other error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: "any other error\n",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			auth := NewAuthorizer(&tokenInspectorWithError{returnedErr: tC.returnedErr}, &Config{})
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Add("Authorization", "Bearer bearer")
			_, b := auth.Authorize(r, w)

			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if b {
				t.Errorf("Expected error in authentication, but was succesful with code %d and body %v", resp.StatusCode, string(body))
			}

			if got, ex := w.Code, tC.expectedCode; got != ex {
				t.Errorf("Expected status code %d, got %d", ex, got)
			}

			if string(body) != tC.expectedBody {
				t.Errorf("Expected body %q, got %q", string(body), tC.expectedBody)
			}
		})
	}
}

func Example() {
	r := mux.NewRouter()
	middleware := Middleware{}

	r.Use(middleware.Handler)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userid, _ := UserID(r.Context())
		log.Printf("AUDIT: User %s does something", userid)

		if HasScope(r.Context(), "dtc:codes:write") {
			_, err := fmt.Fprintf(w, "User has scope.")
			if err != nil {
				panic(err)
			}
			return
		}
		_, err := fmt.Fprintf(w, "Your client may not have the right scopes to see the secret code")
		if err != nil {
			panic(err)
		}
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
		scope:    Scope("scope1 scope2"),
	}

	r := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := security.ContextWithToken(r.Context(), &to)
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

func TestSuccessfulAccessors(t *testing.T) {
	expectedBearerToken := "somevalue"
	expectedUserID := "someuserid"
	expectedClientID := "someclientid"
	expectedBackend := "some-backend"
	expectedScopes := Scope("scope1 scope2")

	var to = token{
		value:    expectedBearerToken,
		userID:   expectedUserID,
		clientID: expectedClientID,
		scope:    expectedScopes,
		backend:  expectedBackend,
	}

	ctx := security.ContextWithToken(context.TODO(), &to)
	newCtx := context.TODO()
	ctx = ContextTransfer(ctx, newCtx)

	uid, _ := UserID(ctx)
	cid, _ := ClientID(ctx)
	backend, _ := Backend(ctx)
	bearerToken, ok := security.GetTokenFromContext(ctx)
	scopes := Scopes(ctx)
	hasScope := HasScope(ctx, "scope2")

	if uid != expectedUserID {
		t.Fatalf("Expected %v, got: %v", expectedUserID, uid)
	}

	if cid != expectedClientID {
		t.Fatalf("Expected %v, got: %v", expectedClientID, cid)
	}

	if backend != expectedBackend {
		t.Fatalf("Expected %v, got: %v", expectedBackend, backend)
	}

	if !ok || bearerToken.GetValue() != expectedBearerToken {
		t.Fatalf("Expected %v, got: %v", expectedBearerToken, bearerToken)
	}

	if scopes[0] != "scope1" || scopes[1] != "scope2" {
		t.Fatalf("Expected %v, got: %v", expectedScopes, scopes)
	}

	if !hasScope {
		t.Fatalf("Expected true, got: %v", hasScope)
	}
}

// Ensure we return sensible results when no data is present, and not panic.
func TestUnsuccessfulAccessors(t *testing.T) {
	ctx := context.TODO()

	uid, uidOK := UserID(ctx)
	cid, cidOK := ClientID(ctx)
	backend, backendOK := Backend(ctx)
	bt, ok := security.GetTokenFromContext(ctx)
	scopes := Scopes(ctx)
	hasScope := HasScope(ctx, "scope2")

	if uid != "" || uidOK {
		t.Fatalf("Expected no %v, got: %v", "UserID", uid)
	}

	if cid != "" || cidOK {
		t.Fatalf("Expected no %v, got: %v", "ClientID", cid)
	}

	if backend != nil || backendOK {
		t.Fatalf("Expected no %v, got: %v", "Backend", backend)
	}

	if ok || bt != nil {
		t.Fatalf("Expected no %v, got: %v", "BearerToken", bt)
	}

	if len(scopes) > 0 {
		t.Fatalf("Expected no scopes, got: %v", scopes)
	}

	if hasScope {
		t.Fatalf("Expected hasScope to return false, got: %v", hasScope)
	}
}

func TestWithBearerToken(t *testing.T) {
	ctx := context.Background()
	ctx = WithBearerToken(ctx, "some access token")
	token, ok := security.GetTokenFromContext(ctx)
	if !ok || token.GetValue() != "some access token" {
		t.Error("could not store bearer token in context")
	}
}
