// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/04 by Vincent Landgraf

// Package http implements a type that helps
// capturing the response status code for use in metrics
package http

import "net/http"

// CaptureStatus exposes the reported StatusCode after
// WriteHeader was called
type CaptureStatus struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader captures the header and the calls the
// next ResponseWriter in the chain
func (r *CaptureStatus) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
