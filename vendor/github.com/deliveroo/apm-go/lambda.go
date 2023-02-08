//go:generate ./bin/mockery --inpackage --name LambdaService --dir .

package apm

import (
	"context"
	"net/http"
	"time"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/lambda"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// LambdaService provides metrics emission service which can be used inside lambdas
type LambdaService interface {
	// Metric sends a distribution metric to DataDog
	Metric(metric string, value float64, tags ...string)
	// MetricWithTimestamp sends a distribution metric to DataDog with a custom timestamp
	MetricWithTimestamp(metric string, value float64, timestamp time.Time, tags ...string)

	// WrapFunction et al - used to initialise the lambda handler
	WrapFunction(handler interface{}) interface{}
	// WrapLambdaHandlerInterface is used to instrument your lambda functions.
	// It returns a modified handler that can be passed directly to the lambda.StartHandler function from aws-lambda-go.
	WrapLambdaHandlerInterface(handler lambda.Handler) lambda.Handler

	// WrapClient modifies the given client's transport to augment it with tracing and returns it.
	WrapClient(c *http.Client) *http.Client
	// WrapRoundTripper returns a new RoundTripper which traces all requests sent over the transport.
	WrapRoundTripper(rt http.RoundTripper) http.RoundTripper

	// StartSpan starts a new span with the given operation name.
	StartSpan(operationName string) ddtrace.Span
	// StartSpanFromContext returns a new span with the given operation name.
	// If a span is found in the context, it will be used as the parent of the resulting span.
	StartSpanFromContext(ctx context.Context, operationName string) (ddtrace.Span, context.Context)
}

type lambdaService struct{}

// NewLambdaService initialises a new service for metrics emission in lambda
func NewLambdaService() LambdaService {
	return &lambdaService{}
}

// WrapFunction is used to instrument your lambda functions.
// It returns a modified handler that can be passed directly to the lambda.Start function from aws-lambda-go.
func (s *lambdaService) WrapFunction(handler interface{}) interface{} {
	return ddlambda.WrapFunction(handler, nil)
}

func (s *lambdaService) WrapLambdaHandlerInterface(handler lambda.Handler) lambda.Handler {
	return ddlambda.WrapLambdaHandlerInterface(handler, nil)
}

// Metric sends a distribution metric to DataDog
func (s *lambdaService) Metric(metric string, value float64, tags ...string) {
	ddlambda.Metric(metric, value, tags...)
}

// MetricWithTimestamp sends a distribution metric to DataDog with a custom timestamp
func (s *lambdaService) MetricWithTimestamp(metric string, value float64, timestamp time.Time, tags ...string) {
	ddlambda.MetricWithTimestamp(metric, value, timestamp, tags...)
}

func (s *lambdaService) WrapClient(c *http.Client) *http.Client {
	return httptrace.WrapClient(c)
}

func (s *lambdaService) WrapRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return httptrace.WrapRoundTripper(rt)
}

// StartSpan starts a new span with the given operation name.
func (s *lambdaService) StartSpan(operationName string) ddtrace.Span {
	return ddtracer.StartSpan(operationName)
}

// StartSpanFromContext returns a new span with the given operation name.
// If a span is found in the context, it will be used as the parent of the resulting span.
func (s *lambdaService) StartSpanFromContext(ctx context.Context, operationName string) (ddtrace.Span, context.Context) {
	return ddtracer.StartSpanFromContext(ctx, operationName)
}
