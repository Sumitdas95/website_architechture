package apm

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/metadata"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type SpanOption func(*Span)

// WithTraceKeep forces DataDog to keep the trace. Use for critical
// transactions, debugging, etc.
func WithTraceKeep() SpanOption {
	return func(s *Span) {
		s.keep = true
	}
}

// WithTraceDrop forces DataDog to drop the trace. Use for uninteresting traces.
func WithTraceDrop() SpanOption {
	return func(s *Span) {
		s.drop = true
	}
}

// WithIncomingRequest inspects the request for a trace id header which should
// have originated from another service. If present, it is set as the span's
// parent id so that DataDog can correlate traces across services.
func WithIncomingRequest(r *http.Request) SpanOption {
	return func(s *Span) {
		s.httpRequest = r
		if tid := r.Header.Get(HTTPHeaderTraceID); tid != "" {
			if i, err := strconv.ParseUint(tid, 10, 64); err == nil {
				s.traceID = i
			}
		}
		if pid := r.Header.Get(HTTPHeaderParentSpanID); pid != "" {
			if i, err := strconv.ParseUint(pid, 10, 64); err == nil {
				s.parentID = i
			}
		}
	}
}

// WithAlertOnError will push any errors to Sentry.
func WithAlertOnError() SpanOption {
	return func(s *Span) {
		s.alertable = true
	}
}

// WithSpanServiceName configures the span with the given service
// name.
func WithSpanServiceName(serviceName string) SpanOption {
	return func(s *Span) {
		s.serviceName = serviceName
	}
}

// WithStartTime sets a custom time as the start time for the created span. By
// default a span is started on span creation.
func WithStartTime(t time.Time) SpanOption {
	return func(s *Span) {
		s.started = t
	}
}

// SpanType is the type of the span.
type SpanType string

const (
	SpanTypeConsumer SpanType = "consumer" // queue operation
	SpanTypeDynamoDB SpanType = "dynamodb" // dynamodb request
	SpanTypeGRPC     SpanType = "grpc"     // gRPC client request
	SpanTypeHTTP     SpanType = "http"     // HTTP client request
	SpanTypeJob      SpanType = "job"      // background job
	SpanTypeMisc     SpanType = "misc"     // miscellaneous
	SpanTypeProducer SpanType = "producer" // queue operation
	SpanTypeRedis    SpanType = "redis"    // redis operation
	SpanTypeRPC      SpanType = "rpc"      // RPC
	SpanTypeSQL      SpanType = "sql"      // sql query
	SpanTypeWeb      SpanType = "web"      // HTTP server request
)

// Span represents a unit of work in a service (e.g. request, task).
//
// See https://github.com/deliveroo/apm-go/issues/20 for the roles of parentID, spanID, and traceID.
type Span struct {
	sync.RWMutex

	alertable   bool      // whether this span should report to Sentry upon an error
	err         error     // terminal error
	finished    time.Time // unit of work finish
	isChild     bool      // whether this span is the child span
	meta        sync.Map  // arbitrary metadata
	name        string    // type of unit of work (e.g. request, task)
	parentID    uint64    // parent trace identifier
	resource    string    // resource (e.g. url path, task name)
	serviceName string    // service (app) identifier
	spanID      uint64    // this span identifier
	spanType    SpanType  // span type
	started     time.Time // unit of work start
	traceID     uint64    // trace identifier

	keep bool // force the DataDog agent to keep the trace
	drop bool // force the DataDog agent to drop the trace

	logger *zap.Logger
	logs   *observer.ObservedLogs // span logs, used in development only

	children []*Span

	tracer *tracer

	grpcMetadata metadata.MD
	httpRequest  *http.Request // request, if a web request span
}

// ID is the tracer identifier for the span.
func (s *Span) ID() uint64 {
	return s.spanID
}

// TraceID identifies the root span of the entire trace.
func (s *Span) TraceID() uint64 {
	return s.traceID
}

// ParentID is the parent's tracer identifier for the span, if any.
func (s *Span) ParentID() uint64 {
	return s.parentID
}

// Name is the type of unit of work (e.g. request, task).
func (s *Span) Name() string { return s.name }

// Resource is the resource corresponding to the unit of work (e.g. url
// path, task name).
func (s *Span) Resource() string { return s.resource }

// ServiceName returns the name of the service that the span is tracing.
func (s *Span) ServiceName() string { return s.serviceName }

// SpanType is the type of the span (e.g. web, sql, redis).
func (s *Span) SpanType() SpanType { return s.spanType }

// Meta retrieves arbitrary metadata for the span.
func (s *Span) Meta() *sync.Map {
	return &s.meta
}

// GetMeta retrieves arbitrary metadata value for the span.
func (s *Span) GetMeta(key string) (string, bool) {
	meta, ok := s.meta.Load(key)

	if !ok {
		return "", ok
	}

	switch v := meta.(type) {
	case string:
		return v, true
	default:
		// explicitly return false in this situation, the user would have to
		// indirectly pull the entire sync.Map to write a non-string value
		// so this is very unlikely but should be treated as a bad case.
		return "", false
	}
}

// Err returns the error (if any) that occurred while completing this
// unit of work.
func (s *Span) Err() error {
	if s == nil {
		return nil
	}

	return s.err
}

// SetError sets the error (if any).
// This function is useful when the span is started and finished outside the function setting the error.
// e.g.: gorillatrace.TracingWithStatusError.
// For everything else prefer using FinishWithAlert, FinishWithError or FinishDeferred.
//
// Errors not passing the filters configured using WithErrorFilters will be ignored.
func (s *Span) SetError(err error) {
	t := s.getTracer()
	if t == nil && s == nil {
		return
	}

	if err != nil && t != nil {
		for _, f := range t.errFilters {
			if f(err) {
				return
			}
		}
	}

	// If the span is nil, we want to make sure the error doesn't get lost.
	if s == nil {
		err = fmt.Errorf("could not associate error to nil span: %w", err)
		t.captureException(0, nil, err, nil)
		if t.wrappedLogger != nil {
			t.wrappedLogger.Warn("could not associate error to nil span", zap.Error(err))
		}
		return
	}

	s.err = err
}

// Started returns the start time of the span.
func (s *Span) Started() time.Time { return s.started }

// Finished returns the finish time of the span. It returns the zero
// time if the span is still in-flight.
func (s *Span) Finished() time.Time { return s.finished }

// Children a copy of the Span's children slice.
func (s *Span) Children() []*Span {
	if s == nil {
		return nil
	}

	s.Lock()
	defer s.Unlock()

	children := make([]*Span, len(s.children))
	copy(children, s.children)
	return children
}

// HTTPRequest returns the http request from context if it exists.
func (s *Span) HTTPRequest() *http.Request {
	return s.httpRequest
}

// Alert records a non-final exception.
func (s *Span) Alert(err error, user *sentry.User, fields ...zap.Field) {
	if err == nil {
		return
	}
	s.alert(3, err, user, fields...)
}

func (s *Span) alert(skipCallers int, err error, user *sentry.User, fields ...zap.Field) {
	s.getTracer().captureException(skipCallers, s, err, user, fields...)
}

// Logger returns the span's logger.
func (s *Span) Logger() *zap.Logger {
	if s == nil {
		return Logger()
	}
	t := s.getTracer()
	s.Lock()
	defer s.Unlock()

	if s.logger == nil {
		if t.isDevelopment {
			coreLogger, logs := observer.New(zap.DebugLevel)
			s.logger = zap.New(coreLogger)
			s.logs = logs
		} else {
			s.logger = t.getLogger(2).With(s.traceLogFields(nil)...)
		}
	}
	return s.logger
}

// SetSQL sets the sql query on the span.
func (s *Span) SetSQL(sql string) {
	s.SetTag(ext.SQLQuery, sql)
}

// SetHTTPURL sets the HTTP url as a meta tag.
func (s *Span) SetHTTPURL(url string) {
	s.SetTag(ext.HTTPURL, url)
}

// SetHTTPMethod sets the HTTP method as a meta tag.
func (s *Span) SetHTTPMethod(method string) {
	s.SetTag(ext.HTTPMethod, method)
}

// SetHTTPStatusCode sets the HTTP status code as a meta tag.
func (s *Span) SetHTTPStatusCode(status string) {
	s.SetTag(ext.HTTPCode, status)
}

// SetTag can be used to set arbitrary tags on the underlying span.
//
// Boolean Value: This will result in a true or false value being assigned to
// the metadata. Can be used to control the AnalyticsEvent or ManualDrop,
// ManualKeep flags.
//
// String Value: This will result in the raw string value being assigned to the
// metadata. Can also be used to control the ResourceName.
//
// Stringer Value: If the value implements fmt.Stringer then `.String()` will
// be called directly on the object. Datadog will attempt to recover from a
// panic if you pass a nil pointer.
//
// Numeric Values: byte, float, int, unsigned integer values will be assigned
// as metric tag on the span. These values must be within  (int64(1) << 53) - 1
// both positive and negative.
//
// Other Value: none numeric, string, fmt.Stringer, bool, or error types will
// just be passed through fmt.Sprint.
func (s *Span) SetTag(key string, value any) {
	if s == nil {
		return
	}

	s.meta.Store(key, value)
}

// SetMeta sets arbitrary metadata on the span.
//
// Deprecated: Use SetTag for improved metrics, metadata and control over your
// Datadog spans.
func (s *Span) SetMeta(keyValPairs ...string) {
	if s == nil {
		return
	}

	i := 0
	for i < len(keyValPairs) {
		if i+1 < len(keyValPairs) {
			s.meta.Store(keyValPairs[i], keyValPairs[i+1])
		} else {
			s.getTracer().getLogger(1).Warn("ignored key without a value", zap.String("ignored", keyValPairs[i]))
		}
		i += 2
	}
}

// Finish marks the unit of work complete. If the span has already been
// finished, any subsequent calls will be ignored, making it safe to
// defer a call to Finish and then call FinishWithError if an error
// occurs.
func (s *Span) Finish() {
	s.finish(nil)
}

// FinishDeferred finishes the span with a possible error.
func (s *Span) FinishDeferred(err *error) {
	s.finish(*err)
}

// FinishWithAlert marks the unit of work complete. If the err isn't nil,
// the error will be reported to Sentry.
func (s *Span) FinishWithAlert(err error) {
	if s == nil {
		return
	}
	s.alertable = true
	s.finish(err)
}

// FinishWithError marks the unit of work failed. If err is nil, calling
// this function is equivalent to calling Finish.
func (s *Span) FinishWithError(err error) {
	s.finish(err)
}

// FinishAndReturn marks the unit of work finished and returns the same error
// back. When it's inconvenient to use FinishDeferred, use FinishAndReturn when
// short-circuiting to return an error:
//
//			func() error {
//				span := apm.NewSpan(...)
//				defer span.Finish()
//				if err := doSomething(); err != nil {
//					return span.FinishAndReturn(err)
//				}
//	         return nil
//			}
func (s *Span) FinishAndReturn(err error) error {
	s.finish(err)
	return err
}

// NewSpan creates a new child span.
func (s *Span) NewSpan(name string, resource string, spanType SpanType, options ...SpanOption) *Span {
	if s == nil {
		return DefaultService.(*apmService).root.NewSpan(name, resource, spanType, options...)
	}
	ss := &Span{
		isChild:  true,
		name:     name,
		parentID: s.spanID,
		resource: resource,
		spanID:   random.Uint64(),
		spanType: spanType,
		started:  time.Now(),
		traceID:  s.traceID,
		tracer:   s.tracer,
	}
	for _, option := range options {
		option(ss)
	}
	// If the service name is missing, default to the parent
	// span's configured name.
	if ss.serviceName == "" {
		ss.serviceName = s.serviceName
	}
	s.Lock()
	defer s.Unlock()
	s.children = append(s.children, ss)
	return ss
}

func (s *Span) getTracer() *tracer {
	if s == nil || s.tracer == nil {
		return DefaultService.(*apmService).root
	}
	return s.tracer
}

// child is true if span is not nil and is the child of another span.
func (s *Span) child() bool {
	return s != nil && s.isChild
}

func (s *Span) finish(err error) {
	if s == nil {
		return
	}
	if !s.finished.IsZero() {
		return // ignore duplicate call
	}
	t := s.getTracer()

	s.Lock()
	s.finished = time.Now()
	if s.err == nil {
		s.SetError(err)
	}
	s.Unlock()
	t.finish(s)
}

// logFields returns the span's details as logging fields. It is primarily used
// when span logging is enabled, but also in the case of warnings about
// unfinished child spans.
func (s *Span) logFields() []zap.Field {
	s.RLock()
	defer s.RUnlock()
	fields := make([]zap.Field, 0)
	add := func(key, val string) {
		fields = append(fields, zap.String(key, val))
	}
	add("name", s.name)
	add("resource", s.resource)
	add("type", string(s.spanType))
	if s.finished.IsZero() {
		add("elapsed", "-")
	} else {
		add("elapsed", s.finished.Sub(s.started).String())
	}

	for _, k := range sortedKeys(&s.meta) {
		if val, ok := s.GetMeta(k); ok {
			add(k, val)
		}
	}

	if s.err != nil {
		add("error", s.err.Error())
		fields = append(fields, zap.Bool("alertable", s.alertable))
	}
	return s.traceLogFields(fields)
}

// traceLogFields appends the span's tracing ids to logging fields to allow for
// correlated logs in DataDog.
func (s *Span) traceLogFields(fields []zap.Field) []zap.Field {
	if s == nil || s.getTracer().isDevelopment {
		return fields
	}
	fields = append(fields, zap.Uint64("dd.trace_id", s.traceID))
	fields = append(fields, zap.Uint64("dd.span_id", s.spanID))
	return fields
}

// SetResource allows to override the existing span resource name.
func (s *Span) SetResource(resource string) {
	s.resource = resource
}

func sortedKeys(m *sync.Map) []string {
	var keys []string

	m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})

	sort.Strings(keys)
	return keys
}
