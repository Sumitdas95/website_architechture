package apm

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

type contextKey struct{}

// ContextWithSpan returns a context with the given Span inside.
func ContextWithSpan(ctx context.Context, s *Span) context.Context {
	return context.WithValue(ctx, contextKey{}, s)
}

// SpanFromContext returns the span stored within the given context.
func SpanFromContext(ctx context.Context) *Span {
	span, _ := ctx.Value(contextKey{}).(*Span)
	return span
}

// AlertFromContext reports an error to Sentry using the span within the given
// context. If no span is found, the alert is reported to Sentry without a
// reference span.
func AlertFromContext(ctx context.Context, err error, user *sentry.User, fields ...zap.Field) {
	SpanFromContext(ctx).alert(3, err, user, fields...)
}

// LoggerFromContext provides a receiver with structured logging methods using
// the span found in the context if it exists, or the apm logger otherwise.
func LoggerFromContext(ctx context.Context, apm ...Service) *zap.Logger {
	if len(apm) == 0 {
		apm = []Service{DefaultService}
	}
	service := apm[0]
	if span := SpanFromContext(ctx); span != nil {
		return span.Logger()
	}
	return service.Logger()
}

// NewSpanFromContext creates a new child span within the given context,
// provided that the context already had a span within.
// If it didn't, it returns a nil, which is safe to use.
func NewSpanFromContext(ctx context.Context, name string, resource string, spanType SpanType, options ...SpanOption) (*Span, context.Context) {
	parent := SpanFromContext(ctx)
	if parent == nil {
		return nil, ctx
	}
	span := parent.NewSpan(name, resource, spanType, options...)
	return span, ContextWithSpan(ctx, span)
}
