// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

// Package jsonapi implements the json api related metrics
package jsonapi

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pace/bricks/http/oauth2"
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

	// Collect pace_api_http_size_bytes histogram metric for the request and response.
	// Now we start our counters to count how many bytes are read from the request and
	// to the response body. Once the body is closed (which the server always does,
	// according to the http.Request.Body documentation) we add our readings to the
	// metrics. This is basically a callback after the handler finished.
	// A special case is when the handler did not read the body. In that case our
	// lenCallbackReader counts the length of the rest as well (by reading it).
	r.Body = &lenCallbackReader{
		reader: r.Body,
		onEOF: func(size int) {
			AddPaceAPIHTTPSizeBytes(float64(size), r.Method, path, serviceName, TypeRequest)
			AddPaceAPIHTTPSizeBytes(float64(m.sizeWritten), r.Method, path, serviceName, TypeResponse)
		},
	}

	return &m
}

// WriteHeader captures the status code for metric submission and
// collects the pace_api_http_request_total counter and
// pace_api_http_request_duration_seconds histogram metric.
func (m *Metric) WriteHeader(statusCode int) {
	clientID, _ := oauth2.ClientID(m.request.Context())
	IncPaceAPIHTTPRequestTotal(strconv.Itoa(statusCode), m.request.Method, m.path, m.serviceName, clientID)

	duration := float64(time.Since(m.requestStart).Nanoseconds()) / float64(time.Second)
	AddPaceAPIHTTPRequestDurationSeconds(duration, m.request.Method, m.path, m.serviceName)
	m.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the length of the response body.
func (m *Metric) Write(p []byte) (int, error) {
	size, err := m.ResponseWriter.Write(p)
	m.sizeWritten += size

	return size, err
}

// IncPaceAPIHTTPRequestTotal increments the pace_api_http_request_total counter metric.
func IncPaceAPIHTTPRequestTotal(code, method, path, service, clientID string) {
	paceAPIHTTPRequestTotal.With(prometheus.Labels{
		"code":      code,
		"method":    method,
		"path":      path,
		"service":   service,
		"client_id": clientID,
	}).Inc()
}

// AddPaceAPIHTTPRequestDurationSeconds adds an observed value for the pace_api_http_request_duration_seconds histogram metric.
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

// lenCallbackReader is a reader that reports the total size before closing.
type lenCallbackReader struct {
	reader io.ReadCloser
	size   int
	onEOF  func(int)
}

func (r *lenCallbackReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.size += n

	return n, err
}

func (r *lenCallbackReader) Close() error {
	// read everything left
	n, _ := io.Copy(io.Discard, r.reader)
	r.size += int(n)
	r.onEOF(r.size)

	return r.reader.Close()
}
