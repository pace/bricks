// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/04 by Vincent Landgraf

// Package jsonapi implements the json api related metrics documented here:
// https://lab.jamit.de/pace/web/meta/wikis/concept/metrics#m2-microservice-any-pace-microservice
package jsonapi

import "github.com/prometheus/client_golang/prometheus"

var (
	paceAPIHTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_api_http_request_total",
			Help: "Collects statistics about each microservice endpoint partitioned by code, method, path service and client_id",
		},
		[]string{"code", "method", "path", "service", "client_id"},
	)
	paceAPIHTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "pace_api_http_request_duration_seconds",
			Help: "Collect performance metrics for each API endpoint partitioned by method, path and service",
		},
		[]string{"method", "path", "service"},
	)
)

// IncPaceAPIHTTPRequestTotal increments the pace_api_http_request_total counter metric
func IncPaceAPIHTTPRequestTotal(code, method, path, service, clientID string) {
	paceAPIHTTPRequestTotal.With(prometheus.Labels{
		"code":      code,
		"method":    method,
		"path":      path,
		"service":   service,
		"client_id": clientID,
	}).Inc()
}

// AddPaceAPIHTTPRequestDurationSeconds adds an observed value for the pace_api_http_request_duration_seconds histogram metric
func AddPaceAPIHTTPRequestDurationSeconds(duration float64, method, path, service string) {
	paceAPIHTTPRequestDurationSeconds.With(prometheus.Labels{
		"method":  method,
		"path":    path,
		"service": service,
	}).Observe(duration)
}
