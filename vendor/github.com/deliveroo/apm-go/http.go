package apm

import (
	"context"
	"net/http"
	"strconv"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// KeyType is the type for the Service value added to the context of a wrapped
// http.RoundTripper, allowing calls
type KeyType struct{}

type roundTripper struct {
	base              http.RoundTripper
	disableServiceTag bool
	Service
}

const (
	HTTPHeaderTraceID      = ddtracer.DefaultTraceIDHeader
	HTTPHeaderParentSpanID = ddtracer.DefaultParentIDHeader
)

var (
	// Key is the context key for the Service value in wrapped http.RoundTripper
	// implementations.
	Key KeyType
)

// NewRoundTripper modifies HTTP requests with a tracing RoundTripper in order
// to inject inter-service tracing headers and track request details and errors.
func NewRoundTripper(rt http.RoundTripper, service Service, options ...TransportOption) http.RoundTripper {
	cfg := parseTransportOptions(options)
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &roundTripper{
		base:              rt,
		disableServiceTag: cfg.disableServiceTag,
		Service:           service,
	}
}

// WrapRoundTripper modifies HTTP requests with a tracing RoundTripper in order
// to inject inter-service tracing headers and track request details and errors.
//
// Deprecated: Use NewRoundTripper instead.
func WrapRoundTripper(rt http.RoundTripper, options ...TransportOption) http.RoundTripper {
	return NewRoundTripper(rt, DefaultService, options...)
}

func (rt *roundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	span, ctx := NewSpanFromContext(req.Context(), "http.request", req.Method, SpanTypeHTTP)
	ctx = context.WithValue(ctx, Key, rt.Service)
	defer span.FinishDeferred(&err)
	if span != nil {
		if req.Header.Get(HTTPHeaderTraceID) == "" {
			req.Header.Set(HTTPHeaderTraceID, strconv.FormatUint(span.traceID, 10))
		}
		if span.spanID != 0 && req.Header.Get(HTTPHeaderParentSpanID) == "" {
			req.Header.Set(HTTPHeaderParentSpanID, strconv.FormatUint(span.spanID, 10))
		}
		span.SetTag(ext.HTTPMethod, req.Method)
		span.SetTag(ext.HTTPURL, req.URL.Path)

		if !rt.disableServiceTag {
			span.SetTag(ext.ServiceName, rt.AppName()+"/"+req.URL.Hostname())
		}
	}
	res, err = rt.base.RoundTrip(req.WithContext(ctx))
	if res != nil && res.StatusCode != 0 {
		span.SetTag(ext.HTTPCode, strconv.Itoa(res.StatusCode))
	}
	return res, err
}
