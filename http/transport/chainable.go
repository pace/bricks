// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/11 by Florian Hübsch

package transport

import "net/http"

type ChainableRoundTripper interface {
	// Transport returns the RoundTripper to make HTTP requests
	Transport() http.RoundTripper
	// SetTransport sets the RoundTripper to make HTTP requests
	SetTransport(http.RoundTripper)
	// RoundTrip executes a single HTTP transaction via Transport()
	RoundTrip(*http.Request) (*http.Response, error)
}

type RoundTripperChain struct {
	// RoundTrippers contains chained round trippers that will be executed in the given order
	RoundTrippers []ChainableRoundTripper
	// Transport that makes the HTTP request at the end of the chain
	Transport http.RoundTripper
}

func ChainWithDefaultTransport(rt ...ChainableRoundTripper) *RoundTripperChain {
	return &RoundTripperChain{
		RoundTrippers: rt,
		Transport:     http.DefaultTransport,
	}
}

func (c *RoundTripperChain) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := c.RoundTrippers

	if len(rt) == 0 {
		return c.Transport.RoundTrip(req)
	}

	for i := 0; i < len(rt)-1; i++ {
		rt[i].SetTransport(rt[i+1])
	}
	rt[len(rt)-1].SetTransport(c.Transport)

	return rt[0].RoundTrip(req)
}
