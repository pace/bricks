// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

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
