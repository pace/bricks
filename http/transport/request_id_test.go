// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDRoundTripper(t *testing.T) {
	rt := RequestIDRoundTripper{}
	rt.SetTransport(&transportWithResponse{})

	t.Run("without req_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/foo", nil)
		_, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.Empty(t, req.Header["Request-Id"])
	})

	t.Run("with req_id", func(t *testing.T) {
		ID := "bqprir5mp1o6vaipufsg"

		r := mux.NewRouter()
		r.Use(log.Handler())
		r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, []string{ID}, r.Header["Request-Id"])
			require.Equal(t, ID, log.RequestID(r))
			require.Equal(t, ID, log.RequestIDFromContext(r.Context()))

			r1 := httptest.NewRequest("GET", "/foo", nil)
			r1 = r1.WithContext(r.Context())

			_, err := rt.RoundTrip(r1)
			assert.NoError(t, err)
			assert.Equal(t, []string{ID}, r1.Header["Request-Id"])
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/foo", nil)
		req.Header.Set("Request-Id", ID)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
