# Changes in apm-go

### 1.39.0
- feat: enable support for global level `shard` configuration for meta and metrics

### 1.38.0
- feat: enable dynamic configuration with Datadog Continuous Profiling.
  - https://docs.datadoghq.com/profiler/search_profiles/?code-lang=go#profile-types

## 1.37
### 1.37.0
- feat: replace internal runtime metrics publishing with DataDogs runtime metric publishing.
  - https://docs.datadoghq.com/tracing/metrics/runtime_metrics/go/

## 1.36
### 1.36.0
- feat: add option to disable adding service tag to outgoing transport spans

## 1.35
### 1.35.0
- fix: refactor options to use a new internal config
  - This removes unnecessary fields from the tracer.
  - It removes calls to `log.Fatal` when parsing options.
  - It captures all config issues instead of just catching the first one.

## 1.34
### 1.34.2
- fix: `span.Err()` will now check if the span is nil with safety guard

### 1.34.1
- fix: `span.Children()` should lock to stop race condition from span being created from parent.

## 1.34
### 1.34.0
- feat: pass gRPC caller name in the metadata from the client

## 1.33
### 1.33.1
- fix: `Span.SetError` function should not panic when span is nil.

### 1.33.0
- feat: gorillatrace middleware now accepts custom span options

## 1.32
### 1.32.0
- feat: bump go version and all dependencies, as well as removing usage of deliveroo/assert-go library

## 1.31
### 1.31.1
- fix: don't override span error if set

### 1.31.0
- feat: adds span function `Span.SetErr(error)` to allow setting errors.

## 1.30
### 1.30.0
- feat: add the `clientID` for HTTP Basic requests to spans

## 1.29
### 1.29.0
- feat: add additional lambda methods

## 1.28
### 1.28.2
- fix: update datadog libs to fix broken transitive dependency
### 1.28.1
- fix: set go version metric on startup
### 1.28.0
- feat: update lambda service interface

## 1.27
### 1.27.0
- feat: add lambda datadog wrapper

## 1.26
### 1.26.1
- fix: panic when finishing span with span logging enabled
### 1.26.0
- feat: return `*zap.Logger` directly from functions instead of returning custom `Logging` interface
  - this is a breaking change for users that define interface with the old signature or implement their own instance of apm.Service
    - this is rare use case and the advantages of returning the concrete type outweigh the negatives of making a breaking change
- feat: remove old ruby build setup and update dependencies
  - build dependencies like golangci-lint are no longer tracked in `go.mod` file, what decreases the footprint of the module
  - latest go version used in CI

## 1.25
### 1.25.1
- fix: align changelog tags with CHANGELOG.md file
### 1.25.0
- feat: add service name to spans in gRPC client interceptor
  - this ensures that we get parity with the http interceptor and get trace metrics
  - it also ensures that we always send trace id in the client even if one is not provided in the context

## 1.24
### 1.24.0
- feat: expose Distribution metric submission type
  - Incompatible change: `Distribution(name string, value float64, rate float64, tagPairs ...string)` has been added to the
    `apm.Metrics` interface.

## 1.23
### 1.23.0
- feat: ensure panics during a web request are reported to sentry
  - previously they would make the span finish with an error (whereas now they finish with an alert)

## 1.22
### 1.22.0
- cleanup tests and verify that we are not leaking goroutines

## 1.21
### 1.21.1
- fix: `span.Meta()` now returns an empty string and ok if the key does not exist
  or the value is not of type string.

## 1.21
### 1.21.0
- Breaking change: `span.Meta()` will no longer return `map[string]string` but `*sync.Map`.
- `span.GetMeta(key string) (string, bool)` now exists to safety get a meta value by key concurrently.

## 1.11
### 1.11.0

- New instance-based API with legacy compatibility layer
- Breaking change: `Logging` is no longer a struct but became an interface.
  The only consequence on legacy code is the need to replace any reference
  to `*apm.Logging` by `apm.Logging`.

## 1.10
### 1.10.1

- Docs fix

### 1.10.0

- feat: add datadog event wrapper to allow go clients raise events (#33)

## 1.9
### 1.9.0

- feat: add ability to override resource name (#32)

## 1.8
### 1.8.0

- feat: awstracer integration (#29)
  - added integration for tracing aws clients

## 1.7
### 1.7.0

- feat: make metric namespace configurable (#28)
  - Added WithMetricsNamespace method for configurable metrics namespace
- feat: slack notification (#25)

## 1.6
### 1.6.1

-  fix(datadog):  downgrade dd-trace-go to 1.27.1 (#24)

### 1.6.0

- feat(datadog): add support for gRPC tracing
  - This adds a basic support for gRPC tracing with Unary operations.

## 1.5
### 1.5.1

- fix(datadog): separate span_id and trace_id
  - This introduces the `span_id` attribute to a span and separates it from
    the `trace_id` which is used for distributed tracing.
  - Both the http server middleware and http client round tripper are updated to
    reflect the change accordingly.
  - Fixes: https://github.com/deliveroo/apm-go/issues/20
- docs(readme): fix profiler code sample (#17)
- docs: add pricing note on profiling
  - This patch adds clarification on the cost implications for Continuous Profiling.

### 1.5.0

- feat(datadog): add profiler support
  - This patch adds support for [Datadog's profiler](https://docs.datadoghq.com/tracing/profiler/getting_started/?tab=go).

## 1.4
### 1.4.0

- fix(datadog): ensure child spans have a service name
  - This patch ensures child spans are created with a service name; if
    WithSpanServiceName is not used on child span creation, the service
    name on the root span is used.

## 1.3
### 1.3.1

- fix(datadog): make service names consistent between tracer and spans
  - This patch makes the service names consistent between ddtracer and its
    traced spans. This fixes an issue with the missing version tags
    introduced in https://github.com/deliveroo/apm-go/pull/9 as the
    version tag is not injected if the span's service name is not equal to
    the tracer's configured service name:
    https://github.com/DataDog/dd-trace-go/blob/ec564257e1787bb2470c30137a004d569d273f3d/ddtrace/tracer/tracer.go#L310-L312.
  - `WithServiceName` is deprecated:
    https://github.com/DataDog/dd-trace-go/blob/ec564257e1787bb2470c30137a004d569d273f3d/ddtrace/tracer/option.go#L286-L299.

### 1.3.0

- feat(pgx): add integration
  This patch adds apm-go integration with pgx's logger and a runtime monitor with StatsD.

## 1.2
### 1.2.0

- feat(datadog): add default service and version tags
  - This patch adds service and version tags by default to all StatsD
    metrics and APM traces. Replicates guidance for
    https://docs.datadoghq.com/tracing/version_tracking.

## 1.1
### 1.1.0

- feat(sentry): update sentry-go (0.5.1 => 0.7.0)

## 1.0
### 1.0.15

- feat(handlers): add custom handlers (#7)

### 1.0.14

- ci: add support for semantic-release
  - This patch adds support for automated release tagging via semantic-release.
- fix: lint and add metric log example

### 1.0.13

- cosmetic: minor clean-up
- feat(statsd): receive StatsD metrics via channel
- fix: span read race during logging
- cosmetic: clean-up log messages
- fix: clean-up CircleCI config

### 1.0.12

- integrations: add pgtrace and gorillatrace packages.
- fix: span: fix race condition when accessing span.meta.
  - As per issue https://github.com/deliveroo/apm-go/issues/3 there was a race
    condition when accessing span.meta.
  - This commit adds a ReadLock/ReadUnlock before accessing span.meta in read mode.

### 1.0.11

- fix: explicitly set whether span is child
  - Previously, whether a span was a child was dependent on the presence of
    a parent id. However, parent ids can come from other services and the
    span may not have a local parent. This would prevent the span from
    reporting (because it was waiting for its non-existent parent to finish).

### 1.0.10

- Add shortcuts for setting http tags

### 1.0.9

- mod: update sentry-go
- cosmetic: remove unused function

### 1.0.8

- cosmetic: add comments and use "name::resource" for spans in logs

### 1.0.7

- Don't log tracer start/stop unless WithSpanLogging is enabled
- Add simple benchmarks for reference
- beta
  - Add tests and locks for race conditions

### 1.0.6

- Ensure span logs always output in development
- Add additional startup log fields
- Remove unused testing object
- doc: Minor README revisions

### 1.0.5

- add more log levels

### 1.0.4 (off main branch)

- Add additional startup log fields
- Remove unused testing object

### 1.0.3

- Flatten fields to tags for Sentry

### 1.0.2

- Drop tracer reference on spans
- fix: fields reported on alerts

### 1.0.1

- Add runtime monitoring

### 1.0.0

- First release
