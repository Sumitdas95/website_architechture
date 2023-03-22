package dependencies

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cep21/circuit/v3"

	"github.com/deliveroo/apm-go"
	"github.com/deliveroo/bnt-internal-test-go/internal/config"
	"github.com/deliveroo/bnt-internal-test-go/internal/httpclient"
)

const circuitBreakerNamePrefix = "httpclient.circuit."

type HTTPClientFactory struct {
	circuitManager    *circuit.Manager
	apmService        apm.Service
	defaultCfg        config.Circuit
	defaultHTTPClient *http.Client
}

// NewHTTPClientFactory constructs a factory to create HTTP clients which are fully operable
func NewHTTPClientFactory(defaultCfg config.Circuit, circuitManager *circuit.Manager, apmService apm.Service, defaultHTTPClient *http.Client) HTTPClientFactory {
	return HTTPClientFactory{
		circuitManager:    circuitManager,
		apmService:        apmService,
		defaultCfg:        defaultCfg,
		defaultHTTPClient: defaultHTTPClient,
	}
}

// Create a new HTTP client, wrapped in a Circuit Breaker, and set up with APM tracing.
// circuitBreakerName is the name of the circuit breaker, this name will be
// used in metric reporting and has to be entirely unique to any other circuit
// used in the application.
//
// This applies even more so if in a circuit manager and configuration
// is based on the circuit name.
func (h HTTPClientFactory) Create(circuitBreakerName string, cfg *config.Circuit) (*http.Client, error) {
	if h.circuitManager == nil {
		return nil, errors.New("no CircuitManager was configured in HTTPClientFactory")
	}

	config := h.defaultCfg
	if cfg != nil {
		config = *cfg
	}

	circuitCfg := circuit.Config{
		Execution: circuit.ExecutionConfig{
			Timeout:               time.Duration(config.Timeout),
			MaxConcurrentRequests: int64(config.MaxConcurrentRequests),
		},
	}

	circuitBreaker, err := h.circuitManager.CreateCircuit(circuitBreakerNamePrefix+circuitBreakerName, circuitCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create circuit: %w", err)
	}

	// make sure that we re-create the client, copying the underlying struct, instead of re-using the same struct which can lead to race conditions
	var client http.Client
	if h.defaultHTTPClient != nil {
		client = *h.defaultHTTPClient
	} else {
		client = *http.DefaultClient
	}

	return httpclient.WithMiddleware(&client,
		httpclient.NewCircuitBreaker(circuitBreaker),
		httpclient.Tracing(h.apmService),
	), nil
}
