package apm

import (
	"strings"

	"google.golang.org/grpc/metadata"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var datadogSpanTypeMappings = map[SpanType]string{
	SpanTypeConsumer: ext.SpanTypeMessageConsumer,
	SpanTypeGRPC:     ext.AppTypeRPC,
	SpanTypeProducer: ext.SpanTypeMessageProducer,
	SpanTypeRedis:    ext.SpanTypeRedis,
	SpanTypeRPC:      ext.AppTypeRPC,
	SpanTypeHTTP:     ext.SpanTypeHTTP,
	SpanTypeSQL:      ext.SpanTypeSQL,
	SpanTypeWeb:      ext.SpanTypeWeb,
}

func (t *tracer) reportDatadogSpan(span *Span, parent ddtracer.Span) {
	span.RLock()
	defer span.RUnlock()
	var parentContext ddtrace.SpanContext
	if span.httpRequest != nil {
		if webContext, err := ddtracer.Extract(ddtracer.HTTPHeadersCarrier(span.httpRequest.Header)); err == nil {
			parentContext = webContext
		}
	}
	if span.grpcMetadata != nil {
		if grpcContext, err := ddtracer.Extract(mdCarrier(span.grpcMetadata)); err == nil {
			parentContext = grpcContext
		}
	}
	if parent != nil {
		parentContext = parent.Context()
	}
	startOpts := []ddtracer.StartSpanOption{
		ddtracer.ChildOf(parentContext),
		ddtracer.ResourceName(span.resource),
		ddtracer.ServiceName(span.serviceName),
		ddtracer.StartTime(span.started),
		ddtracer.WithSpanID(span.spanID),
	}
	set := func(k string, v interface{}) {
		startOpts = append(startOpts, ddtracer.Tag(k, v))
	}
	if spanType, ok := datadogSpanTypeMappings[span.spanType]; ok {
		set(ext.SpanType, spanType)
	}

	switch span.spanType {
	case SpanTypeDynamoDB:
		set(ext.ServiceName, t.appName+"/dynamodb")
	case SpanTypeRedis:
		set(ext.ServiceName, t.appName+"/redis")
	case SpanTypeSQL:
		set(ext.ServiceName, t.appName+"/postgres")
	}

	if span.drop {
		set(ext.ManualDrop, true)
	} else if t.traceKeepAll || span.keep {
		set(ext.ManualKeep, true)
	}

	span.meta.Range(func(key, value interface{}) bool {
		startOpts = append(startOpts, ddtracer.Tag(key.(string), value))
		return true
	})

	ddspan := ddtracer.StartSpan(span.name, startOpts...)
	for _, c := range span.children {
		t.reportDatadogSpan(c, ddspan)
	}
	ddspan.Finish(
		ddtracer.WithError(span.err),
		ddtracer.FinishTime(span.finished),
		ddtracer.StackFrames(64, 6),
	)
}

// mdCarrier satisfies tracer.TextMapWriter and tracer.TextMapReader on top
// of gRPC's metadata, allowing it to be used as a span context carrier for
// distributed tracing.
type mdCarrier metadata.MD

var _ ddtracer.TextMapWriter = (*mdCarrier)(nil)
var _ ddtracer.TextMapReader = (*mdCarrier)(nil)

// Get will return the first entry in the metadata at the given key.
func (mdc mdCarrier) Get(key string) string {
	if m := mdc[key]; len(m) > 0 {
		return m[0]
	}
	return ""
}

// Set will add the given value to the values found at key. Key will be lowercased to match
// the metadata implementation.
func (mdc mdCarrier) Set(key, val string) {
	k := strings.ToLower(key) // as per google.golang.org/grpc/metadata/metadata.go
	mdc[k] = append(mdc[k], val)
}

// ForeachKey will iterate over all key/value pairs in the metadata.
func (mdc mdCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range mdc {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
