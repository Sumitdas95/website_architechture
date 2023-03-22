package dependencies

import (
	"time"

	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"

	"github.com/deliveroo/bnt-internal-test-go/internal/config"
)

// newCircuitBreakerManager sets up the circuit breaker manager, allows the setup
// and configuration of default values per circuit breaker.
func newCircuitBreakerManager(cfg config.Config, defaultConfigurations ...circuit.CommandPropertiesConstructor) *circuit.Manager {
	// default configuration for hystrix based on the provided config. You can
	// choose to override these on a circuit breaker level if you wish.
	hystrixCfg := hystrix.Factory{
		ConfigureOpener: hystrix.ConfigureOpener{
			ErrorThresholdPercentage: int64(cfg.Circuit.ErrorPercentThreshold),
			RequestVolumeThreshold:   int64(cfg.Circuit.RequestVolumeThreshold),
		},
		ConfigureCloser: hystrix.ConfigureCloser{
			SleepWindow: time.Duration(cfg.Circuit.SleepWindow),
		},
	}

	circuitProperties := defaultConfigurations
	circuitProperties = append(circuitProperties, hystrixCfg.Configure)

	return &circuit.Manager{
		DefaultCircuitProperties: circuitProperties,
	}
}
