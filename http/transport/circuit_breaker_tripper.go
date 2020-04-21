package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker"
	"golang.org/x/xerrors"
)

// circuitBreakerTripper implements a ChainableRoundTripper with
// the circuit breaker "gobreaker". By default the circuit breaker
// opens if more than 5 consecutive calls to RoundTrip fail while
// it is closed. To change theses settings a gobreaker.Settings
// object can be passed to NewCircuitBreakerTripper(). For more
// information about the specifiable settings please visit:
// https://github.com/sony/gobreaker
//
// To keep track how often the circuit breaker tripped and
// transitioned to the open/half-open state, a prometheus counter
// is added to each newly instantiated circuit breaker.
//
// CAUTION: It is advised to use this RoundTripper as the last used
// RoundTripper before the actual request is triggered. Otherwise
// the circuit breaker might not open because of real connectivity
// problems, but because of other RoundTrippers.
type circuitBreakerTripper struct {
	transport http.RoundTripper
	breaker   *gobreaker.CircuitBreaker
}

func NewDefaultCircuitBreakerTripper(name string) *circuitBreakerTripper {
	return NewCircuitBreakerTripper(gobreaker.Settings{
		Name: name,
	})
}

func NewCircuitBreakerTripper(settings gobreaker.Settings) *circuitBreakerTripper {
	if settings.Name == "" {
		panic("name is mandatory for circuit breaker")
	}

	stateSwitchCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		ConstLabels: prometheus.Labels{"name": settings.Name},
		Name:        "pace_http_circuit_breaker_state_switch_total",
		Help:        "help",
	}, []string{"from", "to"})

	var ok bool
	var are prometheus.AlreadyRegisteredError
	if err := prometheus.Register(stateSwitchCounterVec); xerrors.As(err, &are) {
		stateSwitchCounterVec, ok = are.ExistingCollector.(*prometheus.CounterVec)
		if !ok {
			panic(fmt.Sprintf(`existing "pace_http_circuit_breaker_state_switch_total" collector no CounterVec, but %T`, are.ExistingCollector))
		}
	} else if err != nil {
		panic(err)
	}

	handler := settings.OnStateChange
	settings.OnStateChange = func(s string, from, to gobreaker.State) {
		if handler != nil {
			handler(s, from, to)
		}

		labels := prometheus.Labels{"from": from.String(), "to": to.String()}
		stateSwitchCounterVec.With(labels).Inc()
	}

	return &circuitBreakerTripper{breaker: gobreaker.NewCircuitBreaker(settings)}
}

// Transport returns the RoundTripper to make HTTP requests
func (c *circuitBreakerTripper) Transport() http.RoundTripper {
	return c.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (c *circuitBreakerTripper) SetTransport(rt http.RoundTripper) {
	c.transport = rt
}

// RoundTrip executes a single HTTP transaction via Transport()
func (c *circuitBreakerTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.breaker.Execute(func() (interface{}, error) {
		return c.transport.RoundTrip(req)
	})
	if err != nil {
		switch {
		case errors.Is(err, gobreaker.ErrOpenState):
			// inform the caller about the broken circuit
			return nil, fmt.Errorf("%w: considering host '%s' unreachable", ErrCircuitBroken, req.Host)
		default:
			return nil, err
		}
	}

	return resp.(*http.Response), nil
}
