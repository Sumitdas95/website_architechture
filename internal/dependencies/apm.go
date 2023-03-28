package dependencies

import (
	"fmt"

	"github.com/deliveroo/apm-go"
	"github.com/deliveroo/bnt-internal-test-go/internal/config"
)

func NewAPM(cfg *config.Config, opts ...apm.Option) (apm.Service, error) {
	switch cfg.Hopper.Environment {
	case "development", "setup", "test":
		opts = append(opts,
			apm.WithAppName(cfg.Hopper.AppName),
			apm.WithEnvironment(cfg.Hopper.Environment),
			apm.WithSpanLogging(true),
			apm.WithStatsDLogging(true),
		)
	default:
		var (
			datadogAddr string
			statsdAddr  string
		)
		if cfg.Datadog.Host != "" && cfg.Datadog.TracerPort != 0 {
			datadogAddr = fmt.Sprintf("%s:%d", cfg.Datadog.Host, cfg.Datadog.TracerPort)
		}
		if cfg.Datadog.Host != "" && cfg.Datadog.StatsDPort != 0 {
			statsdAddr = fmt.Sprintf("%s:%d", cfg.Datadog.Host, cfg.Datadog.StatsDPort)
		}
		opts = append(opts,
			apm.WithAppName(cfg.Hopper.AppName),
			apm.WithDataDogAgentAddr(datadogAddr),
			apm.WithEnvironment(cfg.Hopper.Environment),
			apm.WithReleaseID(cfg.Hopper.ReleaseID),
			apm.WithServiceName(cfg.Hopper.ServiceName),
			apm.WithSpanLogging(cfg.Settings.SpanLogging),
			apm.WithStatsDAddr(statsdAddr),
			apm.WithStatsDLogging(cfg.Settings.StatsDLogging),
			apm.WithRuntimeMonitoring(true),
		)
	}
	apmService, err := apm.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise apm: %w", err)
	}

	return apmService, nil
}
