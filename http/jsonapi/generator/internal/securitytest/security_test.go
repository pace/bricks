// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/01/17 by Charlotte Pröller

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
	oauth2Code, profileKeyCode int
}

func (a *testAuthBackend) AuthorizeOAuth2(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	w.WriteHeader(a.oauth2Code)
	return r.Context(), a.oauth2Code == 200
}

func (a *testAuthBackend) AuthorizeProfileKey(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	w.WriteHeader(a.profileKeyCode)
	return r.Context(), a.profileKeyCode == 200
}

func (testAuthBackend) Init(cfgOAuth2 *oauth2.Config, cfgProfileKey *apikey.Config) {
	//NoOp
}

func TestSecurityBothAuthenticationMethods(t *testing.T) {
	authBackend := &testAuthBackend{200, 200}
	router := Router(&testService{}, authBackend)

	// oauth2 OK, profileKey OK
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result := w.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)
	// oauth2 400, profileKey OK
	w = httptest.NewRecorder()
	authBackend.oauth2Code = http.StatusBadRequest
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	// oauth2 400, profileKey 500
	w = httptest.NewRecorder()
	authBackend.profileKeyCode = http.StatusInternalServerError
	authBackend.oauth2Code = http.StatusBadRequest
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	require.Equal(t, http.StatusUnauthorized, result.StatusCode)

	// oauth2 OK, profileKey 400
	w = httptest.NewRecorder()
	authBackend.profileKeyCode = http.StatusBadRequest
	authBackend.oauth2Code = http.StatusOK
	r = httptest.NewRequest("GET", "http://test.de/pay/beta/test", nil)
	router.ServeHTTP(w, r)
	result = w.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)
}
