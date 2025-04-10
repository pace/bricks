// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

// NewDefaultTransportChain returns a transport chain with retry, jaeger and logging support.
// If not explicitly finalized via `Final` it uses `http.DefaultTransport` as finalizer.
func NewDefaultTransportChain() *RoundTripperChain {
	return Chain(
		&ExternalDependencyRoundTripper{},
		NewDefaultRetryRoundTripper(),
		&LoggingRoundTripper{},
		&LocaleRoundTripper{},
		&RequestIDRoundTripper{},
		// Ensure this is always last, in order to get the correct dump
		NewDumpRoundTripperEnv(),
	)
}

// NewDefaultTransportChainWithExternalName returns a transport chain with retry, jaeger and logging support.
// If not explicitly finalized via `Final` it uses `http.DefaultTransport` as finalizer.
// The passed name is recorded as external dependency.
func NewDefaultTransportChainWithExternalName(name string) *RoundTripperChain {
	return Chain(
		&ExternalDependencyRoundTripper{name: name},
		NewDefaultRetryRoundTripper(),
		&LoggingRoundTripper{},
		&LocaleRoundTripper{},
		&RequestIDRoundTripper{},
		// Ensure this is always last, in order to get the correct dump
		NewDumpRoundTripperEnv(),
	)
}
