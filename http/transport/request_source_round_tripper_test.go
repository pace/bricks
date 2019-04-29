// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/04/29 by Marius Neugebauer

package transport

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestRequestSourceRoundTripper(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo", nil)

	rt := RequestSourceRoundTripper{Header: "foobar"}
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foobar"}, req.Header["Request-Source"])
}
