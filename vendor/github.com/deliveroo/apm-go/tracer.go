package apm

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	ddprofiler "gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

// ProfileType represents a type of profile the datadog profiler is able to run.
type ProfileType int

// Following matching the same type as in dd-trace-go.v1/profiler
const (
	// HeapProfile reports memory allocation samples; used to monitor current
	// and historical memory usage, and to check for memory leaks.
	HeapProfile ProfileType = iota
	// CPUProfile determines where a program spends its time while actively consuming
	// CPU cycles (as opposed to while sleeping or waiting for I/O).
	CPUProfile
	// BlockProfile shows where goroutines block waiting on mutex and channel
	// operations. The block profile is not enabled by default and may cause
	// noticeable CPU overhead. We recommend against enabling it, see
	// DefaultBlockRate for more information.
	BlockProfile
	// MutexProfile reports the lock contentions. When you think your CPU is not fully utilized due
	// to a mutex contention, use this profile. Mutex profile is not enabled by default.
	MutexProfile
	// GoroutineProfile reports stack traces of all current goroutines
	GoroutineProfile
	// expGoroutineWaitProfile reports stack traces and wait durations for
	// goroutines that have been waiting or blocked by a syscall for > 1 minute
	// since the last GC. This feature is currently experimental and only
	// available within DD by setting the DD_PROFILING_WAIT_PROFILE env variable.
	//nolint:deadcode, varcheck, unused
	expGoroutineWaitProfile
	// MetricsProfile reports top-line metrics associated with user-specified profiles
	MetricsProfile
)

// tracer manages tracing for an application.
type tracer struct {
	ctx              context.Context
	cancel           context.CancelFunc
	name             string
	appName          string
	isDevelopment    bool
	handlers         []func(*Span)
	errFilters       []func(error) bool
	logger           *zap.Logger
	wrappedLogger    *zap.Logger
	logSpans         bool
	logStatsD        bool
	sentry           *sentry.Client
	statsd           statsd.ClientInterface
	statsdTags       []string
	statsdChannel    chan StatsDMetric
	statsdNamePrefix string
	traceKeepAll     bool
}

func newTracer(options ...Option) (*tracer, error) {
	cfg, err := parseConfig(options...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	t := &tracer{
		ctx:              ctx,
		cancel:           cancel,
		name:             cfg.name,
		appName:          cfg.appName,
		isDevelopment:    cfg.isDevelopment(),
		handlers:         cfg.handlers,
		errFilters:       cfg.errFilters,
		logSpans:         cfg.logSpans,
		logStatsD:        cfg.logStatsD,
		traceKeepAll:     cfg.traceKeepAll,
		statsdTags:       cfg.statsdTags,
		statsdChannel:    cfg.statsdChannel,
		statsdNamePrefix: cfg.statsdNamePrefix,
	}

	if err := t.setupLogging(cfg); err != nil {
		return nil, err
	}

	if err := t.setupStatsD(cfg); err != nil {
		return nil, err
	}

	setupDataDogAPM(cfg)

	if err := t.setupSentry(cfg); err != nil {
		return nil, err
	}

	if cfg.logSpans {
		t.getLogger(2).
			Info("apm starting",
				zap.String("app-name", cfg.appName),
				zap.String("datadog-agent-addr", cfg.datadogAgentAddr),
				zap.String("datadog-statsd-addr", cfg.datadogStatsDAddr),
				zap.String("env", cfg.env),
				zap.Bool("log-spans", cfg.logSpans),
				zap.Bool("log-statsd", cfg.logStatsD),
				zap.String("name", cfg.name),
				zap.Uint("release-id", cfg.releaseID),
				zap.Bool("runtime-monitor", cfg.runtimeMonitor),
				zap.String("sentry-dsn", cfg.sentryDSN),
				zap.String("service-name", cfg.serviceName),
				zap.String("statsd-addr", cfg.statsdAddr),
				zap.Bool("trace-keep-all", cfg.traceKeepAll),
			)
	}

	t.warnOnVerbosity(cfg)
	if err := setupDataDogProfiling(cfg); err != nil {
		return nil, err
	}

	return t, nil
}

func setupDataDogAPM(cfg *config) {
	if cfg.datadogAgentAddr == "" {
		return
	}
	startOpts := []ddtracer.StartOption{
		ddtracer.WithAgentAddr(cfg.datadogAgentAddr),
		ddtracer.WithDogstatsdAddress(cfg.datadogStatsDAddr),
		ddtracer.WithService(cfg.name),
	}
	if cfg.env != "" {
		startOpts = append(startOpts, ddtracer.WithEnv(cfg.env))
	}
	if cfg.shardName != "" {
		startOpts = append(startOpts, ddtracer.WithGlobalTag("shard", cfg.shardName))
	}
	if cfg.releaseID != 0 {
		startOpts = append(startOpts, ddtracer.WithServiceVersion(strconv.Itoa(int(cfg.releaseID))))
	}
	if cfg.runtimeMonitor {
		// If enabled go and use DataDogs runtime metrics listings.
		// More information can be found [here] which contains metrics
		// designed for their dashboard.
		//
		// [here]: https://docs.datadoghq.com/tracing/metrics/runtime_metrics/go/
		startOpts = append(startOpts, ddtracer.WithRuntimeMetrics())
	}
	ddtracer.Start(startOpts...)
}

func setupDataDogProfiling(cfg *config) error {
	if !cfg.profilingEnabled {
		return nil
	}

	profilerOptions := []ddprofiler.Option{
		ddprofiler.WithService(cfg.name),
		ddprofiler.WithAgentAddr(cfg.datadogAgentAddr),
	}
	// On DD, the default ones are MetricsProfile, CPUProfile, HeapProfile
	if len(cfg.profileTypes) != 0 {
		converted := make([]ddprofiler.ProfileType, len(cfg.profileTypes))
		for i, v := range cfg.profileTypes {
			converted[i] = (ddprofiler.ProfileType)(v)
		}
		profilerOptions = append(profilerOptions, ddprofiler.WithProfileTypes(
			converted...,
		))
	}
	if cfg.env != "" {
		profilerOptions = append(profilerOptions, ddprofiler.WithEnv(cfg.env))
	}
	if cfg.releaseID != 0 {
		profilerOptions = append(profilerOptions, ddprofiler.WithVersion(strconv.Itoa(int(cfg.releaseID))))
	}

	if err := ddprofiler.Start(profilerOptions...); err != nil {
		return err
	}
	return nil
}

// NewSpan creates an isolated span, with no parent and itself as the trace root.
func (t *tracer) NewSpan(name string, resource string, spanType SpanType, options ...SpanOption) *Span {
	spanID := random.Uint64()
	s := &Span{
		name:        name,
		resource:    resource,
		serviceName: t.name,
		spanID:      spanID,
		spanType:    spanType,
		started:     time.Now(),
		traceID:     spanID,
		tracer:      t,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// setupLogging configures logger and logging. If logger is nil, logging will be too.
func (t *tracer) setupLogging(cfg *config) error {
	var (
		err    error
		logger *zap.Logger
	)
	if cfg.logger == nil {
		if cfg.isDevelopment() {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}
		if err != nil {
			return err
		}
		cfg.logger = logger.Named(cfg.name)
	}
	t.logger = cfg.logger
	t.wrappedLogger = cfg.logger.WithOptions(zap.AddCallerSkip(2))
	return nil
}

func (t *tracer) setupStatsD(cfg *config) error {
	if cfg.statsdAddr == "" {
		t.statsd = &statsd.NoOpClient{}
		return nil
	}

	client, err := statsd.New(cfg.statsdAddr)
	if err != nil {
		return err
	}
	if cfg.env != "" {
		client.Tags = append(client.Tags, t.statsdTags...)
	}

	// Set go version before setting the namespace, so that we get a metric without prefix.
	_ = client.Distribution("runtime.go.version", 1, []string{
		"go_version:" + runtime.Version(),
		"app:" + t.appName,
		"service:" + cfg.serviceName,
	}, 1)

	client.Namespace = t.statsdNamePrefix
	t.statsd = client

	return nil
}

func (t *tracer) setupSentry(cfg *config) error {
	if cfg.sentryDSN == "" {
		return nil
	}
	if cfg.env == "" || cfg.releaseID == 0 {
		return errors.New("environment and release id are required for sentry")
	}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         cfg.sentryDSN,
		Environment: cfg.env,
		Release:     strconv.Itoa(int(cfg.releaseID)),
		ServerName:  cfg.appName,
	})
	if err != nil {
		return err
	}
	t.sentry = client
	return nil
}

func (t *tracer) warnOnVerbosity(cfg *config) {
	if t.isDevelopment {
		return
	}
	warn := t.getLogger(2).Warn
	if cfg.logSpans {
		warn("WithSpanLogging is not advised in non-development environments")
	}
	if cfg.logStatsD {
		warn("WithStatsDLogging is not advised in non-development environments")
	}
	if cfg.traceKeepAll {
		warn("WithTraceKeepAll is not advised in non-development environments")
	}
}

func (t *tracer) Stop() error {
	if t == nil {
		// No need to fail: the semantics say the instance cannot be used after Close anyway.
		return nil
	}
	if t.logSpans {
		t.getLogger(2).Info("apm stopping")
	}
	t.Flush(5 * time.Second)
	ddtracer.Stop()
	ddprofiler.Stop()
	t.cancel()
	return nil
}

func (t *tracer) Flush(timeout time.Duration) {
	_ = t.getLogger(1).Sync()
	_ = t.statsd.Flush()
	if t.sentry != nil {
		_ = t.sentry.Flush(timeout)
	}
}

// finish runs the finish handlers for the span.
func (t *tracer) finish(s *Span) {
	if s.child() {
		return
	}
	for _, h := range t.handlers {
		h(s)
	}
	t.reportLogSpan(0, s)
	t.reportSentrySpan(0, s)
	t.reportDatadogSpan(s, nil)
}

func (t *tracer) getLogger(skipCallers int) *zap.Logger {
	if t == nil {
		return nil
	}
	return t.logger.WithOptions(zap.AddCallerSkip(skipCallers))
}

func (t *tracer) reportLogSpan(level int, span *Span) {
	span.RLock()
	defer span.RUnlock()
	if t != nil && t.logSpans {
		prefix := ""
		indent := ""
		if level > 0 {
			indent = strings.Repeat(" ", level-1)
			prefix = indent + "└ "
		}
		logger := t.getLogger(4 + level)
		logger.Info(prefix+span.name+"::"+span.resource, span.logFields()...)
		if span.logs != nil {
			for _, logRow := range span.logs.TakeAll() {
				msg := indent + "| " + logRow.Message
				if ce := logger.Check(logRow.Level, msg); ce != nil {
					ce.Write(logRow.Context...)
				}
			}
		}
	}
}
