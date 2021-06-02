// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/05/10 by Alessandro Miceli

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResponseClientID_Middleare(t *testing.T) {
	AddClientIDToResponse(context.TODO()) // should not fail

	t.Run("empty set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h := ResponseClientID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))
		h.ServeHTTP(rec, req)
		assert.Nil(t, rec.Result().Header[http.CanonicalHeaderKey(ClientIDHeaderName)])
	})
}

func Test_ResponseClientIDContext_String(t *testing.T) {
	var rcc ResponseClientIDContext

	// empty
	assert.Empty(t, rcc.clientID)

	// add client
	rcc.AddClientID("clientID")
	assert.EqualValues(t, "clientID", rcc.clientID)
}
