package wire

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

// FromWire returns a span context from the request if there was an encoded request
func FromWire(r *http.Request) (opentracing.SpanContext, error) {
	return opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
}