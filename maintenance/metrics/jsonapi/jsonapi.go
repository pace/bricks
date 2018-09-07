// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/04 by Vincent Landgraf

// Package jsonapi implements the json api related metrics documented here:
// https://lab.jamit.de/pace/web/meta/wikis/concept/metrics#m2-microservice-any-pace-microservice
package jsonapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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

// Metric is an http.ResponseWriter implementing metrics collector
// because the metrics depend on the http StatusCode.
type Metric struct {
	serviceName string
	path        string // path is the patten path (not the request path)
	http.ResponseWriter
	request      *http.Request
	requestStart time.Time
}

// NewMetric creates a new metric collector (per request) with given
// service and path (pattern! not the request path)
func NewMetric(serviceName, path string, w http.ResponseWriter, r *http.Request) *Metric {
	return &Metric{
		serviceName:    serviceName,
		path:           path,
		ResponseWriter: w,
		request:        r,
		requestStart:   time.Now(),
	}
}

// WriteHeader captures the status code for metric submission and
// collects the pace_api_http_request_total counter and
// pace_api_http_request_duration_seconds histogram metric
func (m *Metric) WriteHeader(statusCode int) {
	// TODO(vil): when oauth2 package is ready, decode clientID from request
	clientID := "none"
	IncPaceAPIHTTPRequestTotal(strconv.Itoa(statusCode), m.request.Method, m.path, m.serviceName, clientID)
	duration := float64(time.Now().Sub(m.requestStart).Nanoseconds()) / float64(time.Second)
	AddPaceAPIHTTPRequestDurationSeconds(duration, m.request.Method, m.path, m.serviceName)
	m.ResponseWriter.WriteHeader(statusCode)
}

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
