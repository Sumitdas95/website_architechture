package circuitbreaker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cep21/circuit/v3"
)

type roundTripper struct {
	inner   http.RoundTripper
	circuit *circuit.Circuit
}

// WrapRoundTripper wraps the provided http.RoundTripper with circuit breaking.
func WrapRoundTripper(rt http.RoundTripper, circuitBreaker *circuit.Circuit) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &roundTripper{inner: rt, circuit: circuitBreaker}
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response

	err := r.circuit.Execute(req.Context(), func(ctx context.Context) (err error) {
		resp, err = r.inner.RoundTrip(req) //nolint:bodyclose // body should be closed by the caller
		return err                         //nolint:wrapcheck // error is wrapped below
	}, nil)
	if err != nil {
		return resp, fmt.Errorf("failed to perform roundtrip using circuitbreaker: %w", err)
	}

	return resp, nil
}
