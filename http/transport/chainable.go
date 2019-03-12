// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/11 by Florian Hübsch

package transport

import "net/http"

// ChainableRoundTripper models a chainable round tripper
type ChainableRoundTripper interface {
	// Transport returns the RoundTripper to make HTTP requests
	Transport() http.RoundTripper
	// SetTransport sets the RoundTripper to make HTTP requests
	SetTransport(http.RoundTripper)
	// RoundTrip executes a single HTTP transaction via Transport()
	RoundTrip(*http.Request) (*http.Response, error)
}

// RoundTripperChain chains multiple chainable round trippers together.
type RoundTripperChain struct {
	// RoundTrippers contains chained round trippers that will be executed in the given order
	RoundTrippers []ChainableRoundTripper
	// Transport that makes the HTTP request at the end of the chain
	Transport http.RoundTripper
}

// Chain returns a round tripper chain with the specified chainable round trippers and http.DefaultTransport as transport.
// The transport can be overriden by using the Final method.
func Chain(rt ...ChainableRoundTripper) *RoundTripperChain {
	return &RoundTripperChain{RoundTrippers: rt, Transport: http.DefaultTransport}
}

// Use adds a chainable round tripper to the round tripper chain.
// It returns the updated round tripper chain.
func (c *RoundTripperChain) Use(rt ChainableRoundTripper) *RoundTripperChain {
	c.RoundTrippers = append(c.RoundTrippers, rt)
	return c
}

// Final sets the transport of the round tripper chain, which is used to make the actual request.
// Final should be called at the end of the chain. If not called, http.DefaultTransport is used.
// It returns the finalized round tripper chain.
func (c *RoundTripperChain) Final(t http.RoundTripper) *RoundTripperChain {
	c.Transport = t
	return c
}

// RoundTrip calls all round trippers in the chain before executing the request.
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
