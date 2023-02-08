package apm

import (
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type config struct {
	name              string
	appName           string
	datadogAgentAddr  string
	datadogStatsDAddr string
	env               string
	shardName         string
	handlers          []func(*Span)
	errFilters        []func(error) bool
	logger            *zap.Logger
	logSpans          bool
	logStatsD         bool
	profilingEnabled  bool
	profileTypes      []ProfileType
	releaseID         uint
	runtimeMonitor    bool
	sentryDSN         string
	serviceName       string
	statsdTags        []string
	statsdAddr        string
	statsdChannel     chan StatsDMetric
	statsdNamePrefix  string
	traceKeepAll      bool
}

func parseConfig(options ...Option) (*config, error) {
	var (
		cfg    config
		outErr error
	)
	for _, opt := range options {
		outErr = multierr.Append(outErr, opt(&cfg))
	}
	if cfg.appName == "" {
		outErr = multierr.Append(outErr, errors.New("application name cannot be blank"))
	}
	if outErr != nil {
		return nil, fmt.Errorf("failed to parse tracer config: %w", outErr)
	}
	if cfg.name == "" {
		cfg.name = cfg.appName
		if cfg.serviceName != "" {
			cfg.name += "-" + cfg.serviceName
		}
	}

	// Include the shard name if and only if the shard was provided.
	// No defaulting to global as this can be misleading.
	if cfg.shardName != "" {
		cfg.statsdTags = append(cfg.statsdTags, "shard:"+cfg.shardName)
	}

	// Do these regardless of statsdAddr: functions don't just send to the
	// client (possibly NoOpClient), but may also emit on the StatsdChannel,
	// so they will use these fields.
	if cfg.statsdNamePrefix == "" {
		cfg.statsdNamePrefix = cfg.name
	}
	cfg.statsdNamePrefix += "."
	cfg.statsdTags = append(cfg.statsdTags, []string{
		"env:" + cfg.env,
		"version:" + strconv.Itoa(int(cfg.releaseID)),
	}...)

	return &cfg, nil
}

func (cfg *config) isDevelopment() bool {
	return cfg.env == DevEnvironment
}
