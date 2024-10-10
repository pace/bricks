// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import "net/http"

// ChainableRoundTripper models a chainable round tripper.
type ChainableRoundTripper interface {
	http.RoundTripper

	// Transport returns the RoundTripper to make HTTP requests
	Transport() http.RoundTripper
	// SetTransport sets the RoundTripper to make HTTP requests
	SetTransport(http.RoundTripper)
}

type finalRoundTripper struct {
	transport http.RoundTripper
}

func (rt *finalRoundTripper) SetTransport(t http.RoundTripper) {
	rt.transport = t
}

func (rt *finalRoundTripper) Transport() http.RoundTripper {
	return rt.transport
}

func (rt *finalRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.transport.RoundTrip(req)
}

// RoundTripperChain chains multiple chainable round trippers together.
type RoundTripperChain struct {
	// first is a pointer to the first chain element
	first ChainableRoundTripper
	// current is a pointer to the current chain element
	current ChainableRoundTripper
	// final is a pointer to the final round tripper (transport)
	final *finalRoundTripper
}

// Chain returns a round tripper chain with the specified chainable round trippers and http.DefaultTransport as transport.
// The transport can be overridden by using the Final method.
func Chain(rt ...ChainableRoundTripper) *RoundTripperChain {
	final := &finalRoundTripper{transport: http.DefaultTransport}
	c := &RoundTripperChain{first: final, current: final, final: final}

	for _, r := range rt {
		c.Use(r)
	}

	return c
}

// Use adds a chainable round tripper to the round tripper chain.
// It returns the updated round tripper chain.
func (c *RoundTripperChain) Use(rt ChainableRoundTripper) *RoundTripperChain {
	// check if chain only consists of final element
	if c.first == c.final {
		c.first = rt
		c.current = rt
		rt.SetTransport(c.final)

		return c
	}

	c.current.SetTransport(rt)
	rt.SetTransport(c.final)

	c.current = rt

	return c
}

// Final sets the transport of the round tripper chain, which is used to make the actual request.
// Final should be called at the end of the chain. If not called, http.DefaultTransport is used.
// It returns the finalized round tripper chain.
func (c *RoundTripperChain) Final(t http.RoundTripper) *RoundTripperChain {
	c.final.SetTransport(t)
	return c
}

// RoundTrip calls all round trippers in the chain before executing the request.
func (c *RoundTripperChain) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.first.RoundTrip(req)
}
