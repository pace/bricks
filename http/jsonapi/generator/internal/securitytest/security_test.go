// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package securitytest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/http/security/apikey"
	"github.com/stretchr/testify/require"
)

type testService struct{}

func (testService) GetTest(ctx context.Context, w GetTestResponseWriter, r *GetTestRequest) error {
	w.OK()
	return nil
}

type testAuthBackend struct {
	oauth2Code, profileKeyCode      int
	canAuthOauth, canAuthProfileKey bool
}

func (a *testAuthBackend) CanAuthorizeOAuth2(r *http.Request) bool {
	return a.canAuthOauth
}

func (a *testAuthBackend) CanAuthorizeProfileKey(r *http.Request) bool {
	return a.canAuthProfileKey
}

func (a *testAuthBackend) AuthorizeOAuth2(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	w.WriteHeader(a.oauth2Code)
	return r.Context(), a.oauth2Code == 200
}

func (a *testAuthBackend) AuthorizeProfileKey(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	w.WriteHeader(a.profileKeyCode)
	return r.Context(), a.profileKeyCode == 200
}

func (testAuthBackend) InitOAuth2(cfgOAuth2 *oauth2.Config) {
	// NoOp
}

func (testAuthBackend) InitProfileKey(cfgProfileKey *apikey.Config) {
	// NoOp
}

func TestSecurityBothAuthenticationMethods(t *testing.T) {
	authBackend := &testAuthBackend{200, 200, true, true}
	router := Router(&testService{}, authBackend)

	// oauth2 OK, profileKey OK, canAuth: both
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result := w.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	// oauth2 ok, profileKey OK, canAuth: none
	authBackend.canAuthProfileKey = false
	authBackend.canAuthOauth = false
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	require.Equal(t, http.StatusUnauthorized, result.StatusCode)

	// oauth2 400, profileKey OK, canAuth = oauth2
	authBackend.canAuthProfileKey = false
	authBackend.canAuthOauth = true
	w = httptest.NewRecorder()
	authBackend.oauth2Code = http.StatusBadRequest
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	require.Equal(t, http.StatusBadRequest, result.StatusCode)

	// oauth2 400, profileKey OK, canAuth = profileKey
	authBackend.canAuthProfileKey = true
	authBackend.canAuthOauth = false
	w = httptest.NewRecorder()
	authBackend.oauth2Code = http.StatusBadRequest
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	// oauth2 400, profileKey 500, canAuth = both
	w = httptest.NewRecorder()
	authBackend.profileKeyCode = http.StatusInternalServerError
	authBackend.oauth2Code = http.StatusBadRequest
	authBackend.canAuthProfileKey = true
	authBackend.canAuthOauth = true
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	// Alphabetic order => get the error of the alphabetic first security scheme
	require.Equal(t, http.StatusBadRequest, result.StatusCode)
}
