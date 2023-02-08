# apm-go

[![CircleCI](https://img.shields.io/circleci/build/github/deliveroo/apm-go?token=20049283e3074c430e96e89fe17556d43955d69f)](https://circleci.com/gh/deliveroo/apm-go/tree/master)
[![GoDoc](https://godoc.org/net/http?status.svg)](https://godoc.deliveroo.net/github.com/deliveroo/apm-go)

Package apm is an opinionated approach to application performance monitoring for
Go services.

It supports distributed tracing (via [DataDog's
APM](https://docs.datadoghq.com/tracing/)), correlated logging (via
[zap-go](https://github.com/uber-go/zap) and DataDog), metrics (via
[StatsD](https://docs.datadoghq.com/developers/dogstatsd/?tab=go)), and error
alerting (via [Sentry](https://sentry.io/)).

It's opinionated in that it requires minimal configuration while providing
strong defaults for development and production environments. It attempts to be
useful out of the box and get out of a developer's way.

It is the successor to [tracer-go](https://github.com/deliveroo/tracer-go), from
which it borrows heavily.

When using `apm-go` throughout your service, DataDog can render clear
visualizations of where your application is spending time by service:

![apm graph by service](./.readme/apm-by-service.png)

Or, by trace type categories:

![apm graph by type](./.readme/apm-by-type.png)

## Initialization

For development, the following configuration is desirable:

```go
func main() {
  apm.Start(
    apm.WithAppName("your-app-name"),
    apm.WithEnvironment(apm.DevEnvironment),
    apm.WithSpanLogging(true),
    apm.WithStatsDLogging(true),
  )
  defer apm.Stop() // should be called before the application exits
}
```

When `WithSpanLogging` is enabled in development, the logger will output tracing
spans (and their logs) along with their children in the following format:

```
INFO    tasks/main.go:38        worker::main     {"name": "worker", "resource": "main", "type": "misc", "elapsed": "67.727671ms", "id": "1"}
INFO    tasks/main.go:38        └ worker::job    {"name": "worker", "resource": "job", "type": "job", "elapsed": "24.089992ms", "job": "2"}
INFO    tasks/main.go:38        └ worker::job    {"name": "worker", "resource": "job", "type": "job", "elapsed": "43.433579ms", "job": "4"}
```

When `WithStatsDLogging` is enabled, the logger will output calls to StatsD in
the logger:

```
INFO    tasks/main.go:22        statsd:gauge   {"name": "jobs.pending", "value": 2, "rate": 1, "tags": []}
```

`WithSpanLogging` and `WithStatsDLogging` shouldn't be necessary to enable in
deployed environments where DataDog and StatsD are available.

For live environments, it's necessary to provide a release identifier, DataDog,
StatsD, and Sentry configuration:

```go
func main() {
  apm.Start(
    apm.WithAppName("your-app-name"),
    apm.WithServiceName("your-service-name"),
    apm.WithEnvironment("production|staging"),
    apm.WithDataDogAgentAddr("..."),
    apm.WithReleaseID("..."),
    apm.WithSentryDSN("..."),
    apm.WithStatsDAddr("..."),
  )
  defer apm.Stop() // should be called before the application exits
}
```

## Tracing spans

Once the apm is started, use it throughout your application for units of work,
such as request handling or background tasks.

Example of a request handler:

```go
type Server struct {}

func (s *Server) handleHello(w http.ResponseWriter, r *http.Request) {
    // Create a new root span.
    span := apm.NewSpan("request", "hello", apm.SpanTypeWeb)
    defer span.Finish()

    // Embed the span within the request context.
    ctx := apm.ContextWithSpan(ctx, span)
    r = r.WithContext(ctx)

    if err := doSomething(ctx); err != nil {
        span.FinishWithError(err)
        fmt.Fprintln(w, "error occurred")
        return
    }

    fmt.Fprintln(w, "Hello, world!")
}
```

Note that it's safe to defer the call to `span.Finish()`, since it will be
ignored if `span.FinishWithError(err)` is called first.

### Child spans

While independent actions like requests and tasks should create root spans,
child actions like database calls and external requests should create child
spans.

Example of a repository that's instrumented its calls using the tracer package:

```go
func (r *Repo) GetFoo(ctx context.Context, id int) (_ *Foo, err error) {
    span := apm.NewSpanFromContext(ctx, "db", "GetFoo", apm.SpanTypeSQL)
    defer span.FinishDeferred(&err)

    // Actually query the database.
}
```

This creates a new child span attached to the root span contained in the
context. `defer span.FinishDeferred(&err)` is a convenient shorthand for:

```go
defer func() {
    if err != nil {
        span.FinishWithError(err)
    } else {
        span.Finish()
    }
}
```

This avoids having to call `span.FinishWithError` at every point where your
function is returning an error.

Finally, sometimes it's just messy or inconvenient to carry an error var
throughout large function that may return at multiple points,
`span.FinishAndReturn` lets you return the error and finish the span in one
line:

```go
func (s *Service) Handler() error {
  // ...
  if err := DoSomething(); err != nil {
    if SafeToIgnoreError(err) {
      return nil
    }
    return span.FinishAndReturn(err)
  }
  if err := DoAnotherThing(); err != nil {
    return span.FinishAndReturn(err)
  }
  // ...
  return nil
}
```

Note that a nil `Span` is safe to use, so you can safely use the above approach
even if you're not certain the context will contain a span.

## Logging

Logging is handled by [zap](https://github.com/uber-go/zap). It provides
structured (reflection-free), leveled logging.

There are several log levels available, `debug`, `info`, `warn`, `error`,
`panic`, `fatal`. Note that the `Logger().Panic` logs then panics and
`Logger().Fatal` logs then does an `os.Exit(1)`. More guidance on log levels is
available in the (zap docs](https://godoc.org/go.uber.org/zap#pkg-constants).

Use the global logger anywhere in your app:

```go
apm.Logger().Debug("hello world")

// zap supports many common field types, see https://godoc.org/go.uber.org/zap#Field
apm.Logger().Info("job starting",
    zap.Int("line-count", lineCount),
    zap.String("file", "2cf06eed8413.csv"),
    zap.Duration("elapsed", duration),
)

// Log level warn and higher will output a stack trace in the logs.
apm.Logger().Warn("oh no", zap.Error(err))
```

In development mode, the log output attempts to be human readable:

```
INFO    main.go:7    job starting    {"line-count": 1388, "file": "2cf06eed8413.csv"}
```

In production mode, logger output is JSON-formatted, which is automatically
parsed by DataDog, making meta fields (facets) viewable and searchable:

```json
{
  "level": "info",
  "ts": 1582230119.672107,
  "caller": "tasks/main.go:7",
  "msg": "job starting",
  "line-count": 1388,
  "file": "2cf06eed8413.csv",
  "elapsed": "1m30s"
}
```

### Correlated logging

Logging directly on a tracing span is a powerful way to debug in deployed
environments. Each log message is associated with the trace, allowing you to
preserve the full context of what occurred without having to search back through
the global stream of log output.

Use the logger on any tracing span and DataDog will associate the log messages
with the span:

```go
func run(s *apm.Span) error {
  span := s.NewSpan("job", "job-name", SpanTypeJob)
  defer span.Finish()
  // ...
  span.Logger().Debug("job params", /* params */)
  // ...
  return nil
}

func handler(ctx context.Context) {
  span := SpanFromContext(ctx)
  span.Logger().Debug("user params",
    zap.String("query", req.Query),
    zap.Int("user-id", userID),
  )
  span.FinishWithError(run(span))
}
```

When the logger is in development mode, logging on a span is prefixed with `|`
at the same level as the parent `Span`:

```
INFO    tasks/main.go:38        worker::main    {"name": "worker", "resource": "main", "type": "misc", "elapsed": "67.292484ms", "id": "1"}
INFO    tasks/main.go:38        └ worker::job   {"name": "worker", "resource": "job", "type": "job", "elapsed": "24.089992ms", "job": "2"}
INFO    tasks/main.go:38        | working       {"duration": "20ms"}
```

You can also use `LoggerFromContext` to either log on the current span or
fallback to the global logger:

```go
func (s *Service) Actuator(ctx context.Context, p Params) error {
  // ...
  logger := apm.LoggerFromContext(ctx)
  logger.Info("actuating",
    zap.String("range", p.Get("range")),
  )
  // ...
  return nil
}
```

## Alerts via Sentry

Too many errors flowing to Sentry by default are noisy and thus less meaningful
or actionable. Package apm encourages using Sentry explicitly via calls like
`apm.Alert(..)` or `span.Alert(..)` in your code base, rather than any time a
trace finishes with an error.

Send an alert to Sentry from anywhere in your code:

```go
apm.Alert(errors.New("unexpected result"), nil,
  zap.Int("index", job.Index),
)
```

There are couple ways of sending alerts from the context of a span. When doing
this, Sentry will receive any meta fields present on the span, along with basic
span details. Use the span option `WithAlertOnError` to alert on all errors that
finish the span:

```go
span := apm.NewSpan("request", "hello", apm.SpanTypeWeb, apm.WithAlertOnError())
span.FinishDeferred(&err)
```

Or, selectively alert on errors within the context of your span:

```go
// ...
defer span.Finish()
// ...
if err != nil {
  span.Alert(err, nil)
}
```

Finally, you can include a `sentry.User` in your alert if this makes sense for
your service. If you provide a user, personally identifiable information like
email and IP address are automatically obfuscated in logs.

```go
apm.SpanFromContext(ctx).Alert(err, &sentry.User{
  ID: user.ID,
  Email: user.Email,
  IPAddress: request.IPAddress,
  Username: request.Username,
})
```

In development mode, alerts are not pushed to Sentry. Instead, you'll see a
`sentry:alert` log message at the `error` level, including a stack trace:

```
ERROR   tasks/main.go:84        sentry:alert    {"index": 1, "sentry.user": {"email": "so************om", "id": "5e4ee951", "ip-address": "1*******1", "username": "Hunter2", "error": "unexpected result"}}
github.com/deliveroo/apm-go.(*tracer).captureException
        /deliveroo/src/apm-go/sentry.go:68
github.com/deliveroo/apm-go.(*tracer).Alert
        /deliveroo/src/apm-go/sentry.go:18
github.com/deliveroo/apm-go.Alert
        /deliveroo/src/apm-go/apm.go:17
main.main
        /deliveroo/src/apm-go/examples/tasks/main.go:84
runtime.main
        /usr/local/Cellar/go/1.13.8/libexec/src/runtime/proc.go:203
```

## StatsD

The full variety of StatsD (and DataDog-specific) metrics are supported via
`apm.StatsD()`. The StatsD client is configured at `apm.Start()` with the
application name concatenated with the service name (e.g.
`orderweb_web_api_orders.`) as the namespace. This can be overridden using
the `WithMetricsNamespace` command. The environment is injected as `env:` tag.

```go
// Count tracks how many times something happened per second
apm.StatsD().Count("orders.created", 100, 1)

// Distribution tracks the statistical distribution of a set of values across your infrastructure
apm.StatsD().Distribution("orders.delivered", 4, 1)

// Gauge measures the value of a metric at a particular time.
apm.StatsD().Gauge("orders.unconfirmed", 2, 1)

// Histogram tracks how many times something happened per second.
apm.StatsD().Histogram("orders.delivered", 4, 1)

// Incr is a Count of 1.
apm.StatsD().Incr("orders.canceled", 1)

// Timing sends timing information in milliseconds.
apm.StatsD().Timing("orders.duration", duration, 1)
```

Adding tags for a metric is done via variadic pairs of strings (rather than
`[]string`):

```go
for _, t := range queue.JobTypes() {
  pending := t.QueueSize()
  apm.StatsD().Gauge("jobs.pending", float64(pending), 1,
    "job-type", t.Name(),
    "job-category", t.Category(),
  )
}
```

In development mode, StatsD calls are visible in log output like so:

```
INFO    tasks/main.go:49        statsd:incr    {"name": "instance.start", "rate": 1, "tags": []}
INFO    tasks/main.go:22        statsd:gauge   {"name": "jobs.pending", "value": 4, "rate": 1, "tags": ["jobtype:delta"]}
```

Sometimes the sheer volume of StatsD calls in a complex application makes
logging them not very useful.

Instead, you can receive all metrics over a channel and then inspect them before
logging. For example, when testing a particular application component you might
filter by name, or metric type:

```go
stats := make(chan apm.StatsDMetric)
go func() {
    for s := range stats {
      // inspect and decide to log a metric or not
    }
}()
apm.Start(
  //...
  apm.WithStatsDChannel(stats),
  //...
)
```

### Runtime monitoring

Use `WithRuntimeMonitoring` at `apm.Start()` to push memory allocator statistics
snapshots to StatsD every ten seconds. These stats can provide insight into
memory and goroutine usage, but will require you to create your a dashboard from
these metrics in DataDog. You can [clone this
dashboard](https://app.datadoghq.com/dashboard/qqk-tzw-pgs/rs-partners-api-franz-runtime)
as a starting point.

```
INFO    apm-go/runtime.go:18    statsd:gauge    {"name": "runtime.goroutines", "value": 2, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:18    statsd:gauge    {"name": "runtime.memory.allocated", "value": 404688, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:18    statsd:gauge    {"name": "runtime.memory.heap", "value": 404688, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:18    statsd:gauge    {"name": "runtime.memory.stack_in_use", "value": 393216, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:19    statsd:count    {"name": "runtime.memory.mallocs", "value": 1557, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:19    statsd:count    {"name": "runtime.memory.frees", "value": 49, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:18    statsd:gauge    {"name": "runtime.memory.live_objects", "value": 1508, "rate": 1, "tags": []}
INFO    apm-go/runtime.go:19    statsd:count    {"name": "runtime.memory.gc_pause", "value": 0, "rate": 1, "tags": []}
```

## Profiling

Use `WithProfiling` at `apm.Start()` to enable Datadog’s
[profiler](https://docs.datadoghq.com/tracing/profiler). Note that
this incurs [additional cost in
Datadog](https://www.datadoghq.com/pricing/?tab=faq-profiler#faq-profiler);
services should make this configurable via an environment variable,
e.g.,

```go
enabled := os.Getenv("ENABLE_PROFILING") == "true"
apm.Start(
	apm.WithProfiler(enabled),
)
```

## External requests and distributed tracing

Distributed tracing is supported when all participating services are configured
to add the appropriate headers to outgoing requests and interpret headers on
incoming requests:

```http
X-Datadog-Trace-Id: 90071992547409920012
X-Datadog-Parent-Id: 60530441770814329149001
```

### HTTP

If you are making requests to other services via HTTP, use `WrapRoundTripper` on
your HTTP client to support distributed tracing on these external requests. The
necessary trace id headers are injected into your request so that DataDog can
correlate the trace.

```go
client := http.DefaultClient
client.Transport = apm.WrapRoundTripper(client.Transport)
```

If you are receiving requests from other services, use a combination of
middleware and the span option `WithIncomingRequest` on your HTTP server to
handle inspection of incoming requests and correlating the trace originating
from another service.

```go
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  span := apm.NewSpan("request", r.Path, apm.SpanTypeWeb,
    WithIncomingRequest(r),
  )
  defer span.Finish()
}

```

### gRPC

If you are making requests to other services via gRPC, use
`UnaryClientInterceptor` on your gRPC client to support distributed tracing on
these external requests. The necessary trace id headers are injected into your
request so that DataDog can correlate the trace.

```go
conn, err := grpc.Dial(addr,
    grpc.WithUnaryInterceptor(apm.UnaryClientInterceptor),
)
```

If you are receiving requests from other services, use `UnaryServerInterceptor`
on your gRPC server to handle inspection of incoming requests and correlating
the trace originating from another service.

```go
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(apm.UnaryServerInterceptor),
)
```

### Slack Alerts
You can send alerts to configured Slack channel.
```go
webhookUrl := "[Your Slack Webhook URL]"

attachment := apm.Attachment{}
attachment.AddField(apm.Field { Title: "Title", Value: "SFTP Error" }).AddField(apm.Field { Title: "Time", Value: time.Now().Format("2006-01-02 15:04:05") })
attachment.AddAction(apm.Action { Type: "button", Text: "Go to DataDog", URL: "https://app.datadoghq.com/logs?query=%40app_name%3Ars-sftp-partners+source%3Aecs-production", Style: "primary" })
attachment.AddAction(apm.Action { Type: "button", Text: "Cancel", URL: "https://app.datadoghq.com/logs?query=%40app_name%3Ars-sftp-partners+source%3Aecs-production", Style: "danger" })
payload := apm.Payload {
    Text: "Failed to connect SFTP server for 30 mins",
    Username: "rs-sftp-partners-worker",
    Channel: "#[Slack Channel Name]",
    Attachments: []apm.Attachment{attachment},
}
err := apm.Send(webhookUrl, "", payload)
```

![slack alert](./.readme/slack-alert.png)

#### Create Slack webhook URL
You can create a webhook URL [here](https://slack.com/intl/en-rs/help/articles/115005265063-Incoming-webhooks-for-Slack)

---

<img src="./.readme/mascot.png" width="150" />
