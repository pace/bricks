// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pace/bricks/http/middleware"
	"github.com/stretchr/testify/assert"
)

type edRoundTripperMock struct {
	req  *http.Request
	resp *http.Response
}

func (m *edRoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	m.req = req
	return m.resp, nil
}

func TestExternalDependencyRoundTripper(t *testing.T) {
	var edc middleware.ExternalDependencyContext
	ctx := middleware.ContextWithExternalDependency(context.Background(), &edc)

	r := httptest.NewRequest("GET", "http://example.com/test", nil)
	r = r.WithContext(ctx)

	mock := &edRoundTripperMock{
		resp: &http.Response{
			Header: http.Header{
				middleware.ExternalDependencyHeaderName: []string{"test1:123,test2:53"},
			},
		},
	}
	lrt := &ExternalDependencyRoundTripper{transport: mock}

	_, err := lrt.RoundTrip(r)
	assert.NoError(t, err)

	assert.EqualValues(t, "test1:123,test2:53", edc.String())
}

func TestExternalDependencyRoundTripperWithName(t *testing.T) {
	var edc middleware.ExternalDependencyContext
	ctx := middleware.ContextWithExternalDependency(context.Background(), &edc)

	r := httptest.NewRequest("GET", "http://example.com/test", nil)
	r = r.WithContext(ctx)

	mock := &edRoundTripperMock{
		resp: &http.Response{
			Header: http.Header{
				middleware.ExternalDependencyHeaderName: []string{"test1:123,test2:53"},
			},
		},
	}
	lrt := &ExternalDependencyRoundTripper{name: "ext", transport: mock}

	_, err := lrt.RoundTrip(r)
	assert.NoError(t, err)

	assert.EqualValues(t, "ext:0,test1:123,test2:53", edc.String())
}
