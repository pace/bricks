// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/05/10 by Alessandro Miceli

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	emptyToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	token      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJhenAiOiJjbGllbnRUZXN0In0.eAUlRLw2R2LEvI9TdaD9P6zGQyz-oF7V-Omm2x00iQk"
)

func TestClientID(t *testing.T) {

	t.Run("empty", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", emptyToken))

		h := ClientID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))
		h.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get(ClientIDHeaderName))
	})

	t.Run("clientID", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

		h := ClientID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))
		h.ServeHTTP(rec, req)
		assert.NotNil(t, rec.Header().Get(ClientIDHeaderName))
	})
}
