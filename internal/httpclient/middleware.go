package httpclient

import (
	"net/http"

	"github.com/cep21/circuit/v3"

	"github.com/deliveroo/apm-go"
	"github.com/deliveroo/test-sonarqube/internal/httpclient/circuitbreaker"
)

type Middleware func(c *http.Client) *http.Client

func WithMiddleware(c *http.Client, middlewares ...Middleware) *http.Client {
	for _, middleware := range middlewares {
		c = middleware(c)
	}
	return c
}

func NewCircuitBreaker(circuitBreaker *circuit.Circuit) Middleware {
	return func(c *http.Client) *http.Client {
		c.Transport = circuitbreaker.WrapRoundTripper(c.Transport, circuitBreaker)
		return c
	}
}

// Tracing is a middleware that enables Datadog APM tracing.
func Tracing(apmService apm.Service) Middleware {
	return func(c *http.Client) *http.Client {
		c.Transport = apm.NewRoundTripper(c.Transport, apmService)
		return c
	}
}
