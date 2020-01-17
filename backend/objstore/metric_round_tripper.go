// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/04/29 by Marius Neugebauer
package objstore

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

type metricRoundTripper struct {
	transport http.RoundTripper
	endpoint  string

	counterTotal  *prometheus.CounterVec
	counterFailed *prometheus.CounterVec
	histogramDur  *prometheus.HistogramVec
}

func (m *metricRoundTripper) Transport() http.RoundTripper {
	return m.transport
}

func (m *metricRoundTripper) SetTransport(rt http.RoundTripper) {
	m.transport = rt
}

func (m *metricRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	labels := prometheus.Labels{
		"method": req.Method,
		"bucket": m.endpoint,
	}

	start := time.Now()
	resp, err := m.Transport().RoundTrip(req)
	dur := time.Since(start)

	// total
	m.counterTotal.With(labels).Inc()

	// duration
	measurable := err != nil
	if measurable {
		// no need to measure timeouts or transport issues
		m.histogramDur.With(labels).Observe(dur.Seconds())
	}

	// failure
	failed := err != nil || m.determineFailure(resp.StatusCode)
	if failed {
		// count transport issues and by status code
		m.counterFailed.With(labels).Inc()
	}

	return resp, err
}

// determineFailure determines whether the response code is considered failed or not.
func (m *metricRoundTripper) determineFailure(code int) bool {
	return !(200 <= code && code < 400)
}
