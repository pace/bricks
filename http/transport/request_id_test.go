// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/maintenance/log"
)

func TestRequestIDRoundTripper(t *testing.T) {
	rt := RequestIDRoundTripper{}
	rt.SetTransport(&transportWithResponse{})

	t.Run("without req_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

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

			r1 := httptest.NewRequest(http.MethodGet, "/foo", nil)
			r1 = r1.WithContext(r.Context())

			resp, err := rt.RoundTrip(r1)
			require.NoError(t, err)

			defer func() {
				err := resp.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, []string{ID}, r1.Header["Request-Id"])
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("Request-Id", ID)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
