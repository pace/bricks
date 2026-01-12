// Copyright © 2018–2026 by PACE Mobility GmbH. All rights reserved.

// Package metric returns the prometheus metrics handler
package metric

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler simply return the prometheus http handler.
// The handler will expose all of the collectors and metrics
// that are attached to the prometheus default registry
func Handler() http.Handler {
	return promhttp.Handler()
}
