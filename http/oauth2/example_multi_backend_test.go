// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package oauth2_test

import (
	"context"
	"fmt"
	"net/http/httptest"

	"github.com/pace/bricks/http/oauth2"
)

var _ oauth2.TokenIntrospecter = (*multiAuthBackends)(nil)

type multiAuthBackends []oauth2.TokenIntrospecter

func (b multiAuthBackends) IntrospectToken(ctx context.Context, token string) (resp *oauth2.IntrospectResponse, err error) {
	for _, backend := range b {
		resp, err = backend.IntrospectToken(ctx, token)
		if resp != nil && err == nil {
			return
		}
	}
	return nil, oauth2.ErrInvalidToken
}

type authBackend [2]string

func (b *authBackend) IntrospectToken(ctx context.Context, token string) (*oauth2.IntrospectResponse, error) {
	if b[1] == token {
		return &oauth2.IntrospectResponse{
			Active:  true,
			Backend: b,
		}, nil
	}
	return nil, oauth2.ErrInvalidToken
}

func Example_multipleBackends() {
	// In case you have multiple authorization backends, you can use the
	// oauth2.Backend(context.Context) function to retrieve the backend that
	// authorized the request. The actual value used for the backend depends on
	// your implementation: you can use constants or pointers, like in this
	// example.

	authorizer := oauth2.NewAuthorizer(multiAuthBackends{
		&authBackend{"A", "token-a"},
		&authBackend{"B", "token-b"},
		&authBackend{"C", "token-c"},
	}, nil)

	r := httptest.NewRequest("GET", "/some/endpoint", nil)
	r.Header.Set("Authorization", "Bearer token-b")

	if authorizer.CanAuthorizeRequest(r) {
		ctx, ok := authorizer.Authorize(r, nil)
		usedBackend, _ := oauth2.Backend(ctx)
		fmt.Printf("%t %s", ok, usedBackend.(*authBackend)[0])
	}

	// Output:
	// true B
}
