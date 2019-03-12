// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/11 by Florian Hübsch

package transport

// DefaultTransport can be used by HTTP clients via the `Transport` field
var DefaultTransport = Chain(&LoggingRoundTripper{})
