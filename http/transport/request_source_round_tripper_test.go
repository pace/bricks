// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestSourceRoundTripper(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	rt := RequestSourceRoundTripper{SourceName: "foobar"}
	rt.SetTransport(&transportWithResponse{})

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	assert.Equal(t, []string{"foobar"}, req.Header["Request-Source"])
}
