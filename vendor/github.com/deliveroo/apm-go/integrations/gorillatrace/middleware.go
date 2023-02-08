package gorillatrace

import (
	"io"
	"net/http"
	"time"

	tracer "github.com/deliveroo/apm-go/integrations/internal/httptrace"

	apmgo "github.com/deliveroo/apm-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

// Tracing is a middleware that wraps each handler in a Span using the provided
// tracer-go Tracer.
// The Span can be recovered within the wrapper handler (to create child spans,
// etc.), through tracer.SpanFromContext(r.Context()).
//
// Deprecated: use TracingWithStatusError instead. TracingWithResponseErrorLogging would ensure that 5xx
// responses are captured as error in the span.
func Tracing(apm apmgo.Service) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			resource := routeName(route)
			rw := tracer.NewTracingResponseWriter(w)
			tracer.TraceAndServe(h, rw, r, resource, apm, false)
		})
	}
}

// TracingWithStatusError is a middleware that wraps each handler in a Span using the provided
// tracer-go Tracer. Using this middleware would ensure that 5xx responses are captured as error in the span.
// The Span can be recovered within the wrapper handler (to create child spans,
// etc.), through tracer.SpanFromContext(r.Context()).
func TracingWithStatusError(apm apmgo.Service, spanOptions ...apmgo.SpanOption) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			resource := routeName(route)
			rw := tracer.NewTracingResponseWriter(w)
			tracer.TraceAndServe(h, rw, r, resource, apm, true, spanOptions...)
		})
	}
}

func routeName(route *mux.Route) string {
	if nil == route {
		return ""
	}
	if n := route.GetName(); n != "" {
		return n
	}
	if n, _ := route.GetPathTemplate(); n != "" {
		return n
	}
	n, _ := route.GetHostTemplate()
	return n
}

// StdLogging is a middleware that logs all requests to the wrapped handler to the
// provided io.Writer in Apache Common Log Format (CLF).
func StdLogging(out io.Writer) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(out, h)
	}
}

// SpanLogging is a middleware that logs all requests to the wrapped handler.
// It uses apm-go span level structured logging, reporting handler-related
// metrics and tags, including response status.
func SpanLogging(apm apmgo.Service) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return handlers.CustomLoggingHandler(nil, h, structuredLogFormatter(apm))
	}
}

func structuredLogFormatter(apm apmgo.Service) handlers.LogFormatter {
	return func(_ io.Writer, params handlers.LogFormatterParams) {
		ctx := params.Request.Context()
		log := apmgo.LoggerFromContext(ctx, apm)

		route := mux.CurrentRoute(params.Request)
		resource := routeName(route)
		zapParams := []zap.Field{
			zap.String("level", "INFO"),
			zap.String(ext.HTTPMethod, params.Request.Method),
			zap.String(ext.HTTPURL, params.URL.Path),
			zap.Duration("request.time_elapsed", time.Since(params.TimeStamp)),
			zap.Int("http.status_code", params.StatusCode),
		}

		if params.StatusCode > 499 {
			log.Error(params.Request.Method+" "+resource, zapParams...)
			return
		}
		log.Info(params.Request.Method+" "+resource, zapParams...)
	}
}
