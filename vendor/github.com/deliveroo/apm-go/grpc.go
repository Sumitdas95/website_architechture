//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpctest/grpctest.proto
package apm

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

const (
	// GRPCMetaCallerAppNameKey is a key used in the gRPC metadata that specifies the app name of the caller.
	GRPCMetaCallerAppNameKey = "roo-caller-app-name"
)

// UnaryClientInterceptor decorates the context by injecting inter-service
// tracing headers and track request details from a new child Span.
//
// Deprecated: Use NewUnaryClientInterceptor instead.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	return NewUnaryClientInterceptor(DefaultService)(ctx, method, req, reply, cc, invoker, opts...)
}

// NewUnaryClientInterceptor decorates the context by injecting inter-service
// tracing headers and track request details from a new child Span.
func NewUnaryClientInterceptor(service Service, options ...TransportOption) grpc.UnaryClientInterceptor {
	cfg := parseTransportOptions(options)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		span, ctx := NewSpanFromContext(ctx, "grpc.client", method, SpanTypeGRPC)
		if span == nil {
			// Initialise span if we didn't get one from the context.
			// This ensures that we can always identify a caller even
			// if we don't get a span in the context.
			span = service.NewSpan("grpc.client", method, SpanTypeGRPC)
		}
		defer span.FinishDeferred(&err)

		if !cfg.disableServiceTag {
			span.SetTag(ext.ServiceName, service.AppName()+method)
		}

		ctx = metadata.AppendToOutgoingContext(ctx,
			HTTPHeaderTraceID, fmt.Sprint(span.TraceID()),
			HTTPHeaderParentSpanID, fmt.Sprint(span.ID()),
			GRPCMetaCallerAppNameKey, service.AppName(),
		)
		var p peer.Peer
		opts = append(opts, grpc.Peer(&p))
		err = invoker(ctx, method, req, reply, cc, opts...)
		setSpanTargetFromPeer(span, p)
		setSpanResponseCodeFromErr(span, err)
		return err
	}
}

// setSpanTargetFromPeer sets the target tags in a span based on the gRPC peer.
func setSpanTargetFromPeer(span *Span, p peer.Peer) {
	// if the peer was set, set the meta tags
	if p.Addr != nil {
		host, port, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			if host != "" {
				span.SetTag(ext.TargetHost, host)
			}
			span.SetTag(ext.TargetPort, port)
		}
	}
}

// UnaryServerInterceptor injects a new Span in the context, referencing the
// parent span if the inter-service tracing headers are present.
//
// Deprecated: Use NewUnaryServerInterceptor instead.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return NewUnaryServerInterceptor(DefaultService)(ctx, req, info, handler)
}

// NewUnaryServerInterceptor injects a new Span in the context, referencing the
// parent span if the inter-service tracing headers are present.
func NewUnaryServerInterceptor(service Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		span := service.NewSpan("grpc.server", info.FullMethod, SpanTypeGRPC, withParentTraceIDFromContext(ctx))
		defer span.FinishDeferred(&err)

		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)

			// Track the distribution of grpc request durations in milliseconds
			// while ensuring to include the `grpc.status_code` .
			service.StatsD().Distribution("roo.grpc.request.latency",
				float64(duration.Milliseconds()), 1, []string{
					"grpc.status_code", status.Code(err).String(),
					"service", span.ServiceName(),
					"resource", span.Resource(),
				}...)
		}()

		span.SetTag("grpc.method.kind", "unary")
		span.SetTag("grpc.method.name", info.FullMethod)
		ctx = ContextWithSpan(ctx, span)

		resp, err = handler(ctx, req)
		setSpanResponseCodeFromErr(span, err)

		return resp, err
	}
}

func setSpanResponseCodeFromErr(span *Span, err error) {
	span.SetTag("grpc.status_code", status.Code(err).String())
}

// withParentTraceIDFromContext inspects the context for a trace id header which
// should have originated from another gRPC service. If present, it is set as
// the span's parent id so that DataDog can correlate traces across services.
func withParentTraceIDFromContext(ctx context.Context) SpanOption {
	return func(s *Span) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return
		}
		if tid := md.Get(HTTPHeaderTraceID); len(tid) > 0 {
			if i, err := strconv.ParseUint(tid[0], 10, 64); err == nil {
				s.traceID = i
			}
		}
		if pid := md.Get(HTTPHeaderParentSpanID); len(pid) > 0 {
			if i, err := strconv.ParseUint(pid[0], 10, 64); err == nil {
				s.parentID = i
			}
		}
		s.grpcMetadata = md
	}
}
