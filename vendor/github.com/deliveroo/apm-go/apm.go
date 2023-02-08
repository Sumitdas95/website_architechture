//go:generate ./bin/mockery --inpackage --name Logging --dir .
//go:generate ./bin/mockery --inpackage --name Metrics --dir .
//go:generate ./bin/mockery --inpackage --name Service --dir .

package apm

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

const (
	DefaultAppName      = "default"
	DevEnvironment      = "development"
	DefaultEnvironment  = DevEnvironment
	DefaultFlushTimeout = 5 * time.Second
)

var (
	DefaultService Service
)

// Service is the entry point to the APM which all packages should use.
type Service interface {
	// Alert reports an error to Sentry if configured.
	// All zap.Field parameters are remapped to Sentry tags.
	Alert(err error, user *sentry.User, fields ...zap.Field)

	// AppName reports on the name the service is performing APM for.
	AppName() string

	// Closer flushes all pending information and releases connections.
	// The service instance may no longer be used afterwards, except for its
	// Close() method, which is idempotent.
	io.Closer

	// Flush flushes any buffered logs, traces, or errors.
	Flush(timeout time.Duration)

	// Logger provides a receiver with the global set of logging functions.
	Logger() *zap.Logger

	// NewSpan creates a new span.
	NewSpan(name string, resource string, spanType SpanType, options ...SpanOption) *Span

	// StatsD provides a receiver for sending metrics like counters and gauges.
	StatsD() Metrics
}

type apmService struct {
	root *tracer
}

// Option is a configuration option for the APM.
type Option func(cfg *config) error

func init() {
	// Do not move to var declaration to avoid an initialization cycle.
	DefaultService = MustNew(
		WithAppName(DefaultAppName),
		WithEnvironment(DefaultEnvironment),
	)
}

// MustNew returns a new Service implementation configured with the passed options.
// To avoid failing in case of error, use New.
func MustNew(option ...Option) Service {
	s, err := New(option...)
	if err != nil {
		log.Fatalln(err)
	}
	return s
}

// New returns a new Service implementation configured with the passed options.
// To fail in case of error, use MustNew.
func New(options ...Option) (Service, error) {
	t, err := newTracer(options...)
	if err != nil {
		return nil, err
	}

	s := &apmService{
		root: t,
	}
	return s, nil
}

func (s *apmService) ensureTracer() {
	if s == nil {
		log.Fatal("apm.Service is not configured")
		return // Useless, but required by golangci-lint to avoid a SA5011.
	}
	if s.root == nil {
		log.Fatal("apm.Service root is not configured")
	}
}

func (s *apmService) Alert(err error, user *sentry.User, fields ...zap.Field) {
	s.ensureTracer()
	s.root.alert(0, err, user, fields...)
}

func (s *apmService) Close() error {
	if s == nil || s.root == nil {
		// No need to fail: the semantics say the instance cannot be used after Close anyway.
		return nil
	}
	stopper := s.root.Stop
	s.root = nil
	return stopper()
}

func (s *apmService) Flush(timeout time.Duration) {
	s.ensureTracer()
	s.root.Flush(timeout)
}

func (s *apmService) Logger() *zap.Logger {
	s.ensureTracer()
	return s.root.wrappedLogger
}

func (s *apmService) AppName() string {
	s.ensureTracer()
	return s.root.appName
}

// NewSpan creates a new span.
func (s *apmService) NewSpan(name string, resource string, spanType SpanType, options ...SpanOption) *Span {
	s.ensureTracer()
	return s.root.NewSpan(name, resource, spanType, options...)
}

func (s *apmService) StatsD() Metrics {
	s.ensureTracer()
	return s.root
}

// Deprecated: use myAPMInstance.Alert().
func Alert(err error, user *sentry.User, fields ...zap.Field) {
	DefaultService.Alert(err, user, fields...)
}

// Deprecated: use myAPMInstance.Flush().
func Flush(timeout time.Duration) {
	DefaultService.Flush(timeout)
}

// Deprecated: use myAPMInstance.Logger()
func Logger() *zap.Logger {
	return DefaultService.Logger()
}

// Deprecated: use myAPMInstance.NewSpan().
func NewSpan(name string, resource string, spanType SpanType, options ...SpanOption) *Span {
	return DefaultService.NewSpan(name, resource, spanType, options...)
}

// Start is the legacy function to initialize the APM. It overwrites the current
// tracer with a blank one.
//
// Deprecated: use New(options...) or MustNew(options...)
func Start(options ...Option) {
	DefaultService = MustNew(options...)
}

// Deprecated: use myAPMInstance.StatsD()
func StatsD() Metrics {
	return DefaultService.StatsD()
}

// Stop is the legacy function to stop the APM. It resets the default instance to nil.
//
// Deprecated: use myAPMInstance.Close() instead.
func Stop() {
	_ = DefaultService.Close()
}

// WithAppName sets the application name used to configure the DataDog APM and
// StatsD reporting.
func WithAppName(name string) Option {
	return func(cfg *config) error {
		if name == "" {
			return errors.New("attempted to configure the tracer with an empty app name")
		}
		cfg.appName = name
		return nil
	}
}

// WithName sets the name used to configure the DataDog APM and
// StatsD reporting.
func WithName(name string) Option {
	return func(cfg *config) error {
		if name == "" {
			return errors.New("attempted to configure the tracer with an empty name")
		}
		cfg.name = name
		return nil
	}
}

// WithLogger sets the logger instance.
func WithLogger(logger *zap.Logger) Option {
	return func(cfg *config) error {
		if logger == nil {
			return errors.New("attempted to configure the tracer with a nil logger")
		}
		cfg.logger = logger
		return nil
	}
}

// WithDataDogStatsDAddress specifies the address to connect to for sending metrics
// to the Datadog Agent.
func WithDataDogStatsDAddress(addr string) Option {
	return func(cfg *config) error {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("attempted to configure invalid DataDatog StatsD address [%s]: %v", addr, err)
		}

		cfg.datadogStatsDAddr = net.JoinHostPort(host, port)
		return nil
	}
}

// WithDataDogAgentAddr sets the address for the DataDog agent.
func WithDataDogAgentAddr(addr string) Option {
	return func(cfg *config) error {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("attempted to configure invalid DataDatog agent address [%s]: %v", addr, err)
		}
		cfg.datadogAgentAddr = net.JoinHostPort(host, port)
		return nil
	}
}

// WithEnvironment sets the environment, e.g. staging or production, which
// configures logging behavior. The default is development.
func WithEnvironment(env string) Option {
	return func(cfg *config) error {
		if env == "" {
			env = DefaultEnvironment
		}
		cfg.env = env
		return nil
	}
}

// WithShard sets the shard name, e.g. global, mle-01, eur-01, etc, which
// configures the global tag set. The default will result in the tag not
// being set.
func WithShard(shardName string) Option {
	return func(cfg *config) error {
		cfg.shardName = shardName
		return nil
	}
}

// WithProfiler enables the Datadog profiler.
//
// See https://docs.datadoghq.com/tracing/#continuous-profiler
func WithProfiler(enabled bool) Option {
	return func(cfg *config) error {
		cfg.profilingEnabled = enabled
		return nil
	}
}

// WithProfileTypes configures the Datadog profiler types.
// See dd-trace-go.v1/profiler/profile.go:26 for available types
func WithProfileTypes(types ...ProfileType) Option {
	return func(cfg *config) error {
		cfg.profileTypes = types
		return nil
	}
}

// WithReleaseID sets the application release ID.
func WithReleaseID(releaseID string) Option {
	return func(cfg *config) error {
		v, err := strconv.Atoi(releaseID)
		if err != nil || v < 0 {
			// Coverage exists but is lost because of the subprocess logic.
			// See https://github.com/golang/go/issues/28235
			return fmt.Errorf("attempting to configure non-positive-integer release [%s]", releaseID)
		}
		cfg.releaseID = uint(v)
		return nil
	}
}

// WithSentryDSN sets the remote address for the Sentry client.
func WithSentryDSN(sentryDSN string) Option {
	return func(cfg *config) error {
		_, err := sentry.NewDsn(sentryDSN)
		if err != nil {
			return fmt.Errorf("attempting to configure invalid Sentry DSN %s: %v", sentryDSN, err)
		}
		cfg.sentryDSN = sentryDSN
		return nil
	}
}

// WithServiceName sets the service name used to configure the APM and StatsD
// reporting.
func WithServiceName(name string) Option {
	return func(cfg *config) error {
		if name == "" {
			return errors.New("attempting to configure empty service name")
		}
		cfg.serviceName = name
		return nil
	}
}

// WithSpanLogging enables span logging. Not advised for use in production.
func WithSpanLogging(enabled bool) Option {
	return func(cfg *config) error {
		cfg.logSpans = enabled
		return nil
	}
}

// WithStatsDLogging enables statsd logging. Not advised for use in production.
func WithStatsDLogging(enabled bool) Option {
	return func(cfg *config) error {
		cfg.logStatsD = enabled
		return nil
	}
}

// WithStatsDChannel allows you to receive all apm StatsD calls over a channel.
//
// It's useful in development environments to log a subset of StatsD calls when
// developing or verify that certain metrics are sent in integration tests. Not
// advised for use in production.
func WithStatsDChannel(c chan StatsDMetric) Option {
	return func(cfg *config) error {
		if c == nil {
			return errors.New("attempting to configure a nil statsd channel")
		}
		cfg.statsdChannel = c
		return nil
	}
}

// WithStatsDAddr sets the address for the DataDog StatsD client.
func WithStatsDAddr(addr string) Option {
	return func(cfg *config) error {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("attempted to configure invalid statsd address %s: %v", addr, err)
		}
		cfg.statsdAddr = net.JoinHostPort(host, port)

		// If someone has not already configured the `datadogStatsDAddr` go and
		// set it up as well. This is used during datadog runtime profiling
		// which creates its own client.
		if cfg.datadogStatsDAddr == "" {
			cfg.datadogStatsDAddr = net.JoinHostPort(host, port)
		}

		return nil
	}
}

// WithRuntimeMonitoring toggles runtime monitoring.
func WithRuntimeMonitoring(enabled bool) Option {
	return func(cfg *config) error {
		cfg.runtimeMonitor = enabled
		return nil
	}
}

// WithTraceKeepAll ensures all traces will be indexed. Not advised for use in
// production.
//
//   - Current link: https://docs.datadoghq.com/tracing/trace_retention_and_ingestion/
//   - Legacy link: https://docs.datadoghq.com/tracing/guide/trace_sampling_and_storage/?tab=go
func WithTraceKeepAll(keep bool) Option {
	return func(cfg *config) error {
		cfg.traceKeepAll = keep
		return nil
	}
}

func WithMetricsNamespace(namespace string) Option {
	return func(cfg *config) error {
		split := strings.Split(namespace, ".")
		for i := 0; i < len(split); i++ {
			if split[i] == "" {
				return fmt.Errorf("attempting to configure invalid metrics namespace (consecutive dots) %s", namespace)
			}
			for _, r := range split[i] {
				if !unicode.In(r, unicode.L) && r != '_' {
					return fmt.Errorf("attempting to configure invalid metrics namespace (invalid character '%c') %s", r, namespace)
				}
			}
		}
		cfg.statsdNamePrefix = namespace
		return nil
	}
}

// WithCustomHandler adds a custom handler, which will be passed the parent Span
// when it finishes, but before it's reported to DataDog, Sentry, and the logger.
func WithCustomHandler(f func(*Span)) Option {
	return func(cfg *config) error {
		if f == nil {
			return errors.New("attempting to configure nil custom handler")
		}
		cfg.handlers = append(cfg.handlers, f)
		return nil
	}
}

// WithErrorFilters ignores span errors that match filters.
func WithErrorFilters(errFilters ...func(error) bool) Option {
	return func(cfg *config) error {
		cfg.errFilters = errFilters
		return nil
	}
}

// WithStatsDTags adds initial set of statsd tags.
func WithStatsDTags(tags []string) Option {
	return func(cfg *config) error {
		cfg.statsdTags = tags
		return nil
	}
}
