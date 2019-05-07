// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/04 by Vincent Landgraf

// Package jsonapi implements the json api related metrics documented here:
// https://lab.jamit.de/pace/web/meta/wikis/concept/metrics#m2-microservice-any-pace-microservice
package jsonapi

import (
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pace/bricks/http/oauth2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	kb = 1024
	mb = kb * kb
)

const (
	TypeRequest  = "req"
	TypeResponse = "resp"
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
	paceAPIHTTPSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "pace_api_http_size_bytes",
			Help: "Collect request and response body size for each API endpoint partitioned by method, path and service",
			Buckets: []float64{
				100, kb, 10 * kb, 100 * kb,
				1 * mb, 5 * mb, 10 * mb, 100 * mb,
			},
		},
		[]string{"method", "path", "service", "type"},
	)
)

func init() {
	prometheus.MustRegister(paceAPIHTTPRequestTotal)
	prometheus.MustRegister(paceAPIHTTPRequestDurationSeconds)
	prometheus.MustRegister(paceAPIHTTPSizeBytes)
}

// Metric is an http.ResponseWriter implementing metrics collector
// because the metrics depend on the http StatusCode.
type Metric struct {
	serviceName string
	path        string // path is the patten path (not the request path)
	http.ResponseWriter
	request      *http.Request
	requestStart time.Time
	sizeWritten  int
}

// NewMetric creates a new metric collector (per request) with given
// service and path (pattern! not the request path) and collects the
// pace_api_http_size_bytes histogram metric.
func NewMetric(serviceName, path string, w http.ResponseWriter, r *http.Request) *Metric {
	m := Metric{
		serviceName:    serviceName,
		path:           path,
		ResponseWriter: w,
		request:        r,
		requestStart:   time.Now(),
	}

	// collect pace_api_http_size_bytes histogram metric for the request and response
	r.Body = &lenCallbackReader{
		r: r.Body,
		onEOF: func(size int) {
			AddPaceAPIHTTPSizeBytes(float64(size), r.Method, path, serviceName, TypeRequest)
			AddPaceAPIHTTPSizeBytes(float64(m.sizeWritten), r.Method, path, serviceName, TypeResponse)
		},
	}

	return &m
}

// WriteHeader captures the status code for metric submission and
// collects the pace_api_http_request_total counter and
// pace_api_http_request_duration_seconds histogram metric
func (m *Metric) WriteHeader(statusCode int) {
	clientID, _ := oauth2.ClientID(m.request.Context())
	IncPaceAPIHTTPRequestTotal(strconv.Itoa(statusCode), m.request.Method, m.path, m.serviceName, clientID)
	duration := float64(time.Since(m.requestStart).Nanoseconds()) / float64(time.Second)
	AddPaceAPIHTTPRequestDurationSeconds(duration, m.request.Method, m.path, m.serviceName)
	m.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the length of the response body.
func (m *Metric) Write(p []byte) (int, error) {
	m.sizeWritten += len(p)
	return m.ResponseWriter.Write(p)
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

// AddPaceAPIHTTPSizeBytes adds an observed value for the pace_api_http_size_bytes histogram metric.
func AddPaceAPIHTTPSizeBytes(size float64, method, path, service, requestOrResponse string) {
	paceAPIHTTPSizeBytes.With(prometheus.Labels{
		"method":  method,
		"path":    path,
		"service": service,
		"type":    requestOrResponse,
	}).Observe(size)
}

// lenCallbackReader is a reader that reports the total size before closing
type lenCallbackReader struct {
	r     io.ReadCloser
	size  int
	onEOF func(int)
}

func (r *lenCallbackReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.size += n
	return n, err
}

func (r *lenCallbackReader) Close() error {
	// read everything left
	n, _ := io.Copy(ioutil.Discard, r.r)
	r.size += int(n)
	r.onEOF(r.size)
	return r.r.Close()
}
