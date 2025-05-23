// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.

package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/pace/bricks/http/middleware"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/log"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"
)

// Deprecated: Use NewClient instead.
func DialContext(_ context.Context, addr string) (*grpc.ClientConn, error) {
	return NewClient(addr)
}

// Deprecated: Use NewClient instead.
func Dial(addr string) (*grpc.ClientConn, error) {
	return NewClient(addr)
}

func NewClient(addr string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn

	clientMetrics := grpc_prometheus.NewClientMetrics()

	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
		grpc_retry.WithMax(10),
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(
			grpc_sentry.StreamClientInterceptor(),
			grpc_retry.StreamClientInterceptor(opts...),
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				start := time.Now()
				cs, err := streamer(prepareClientContext(ctx), desc, cc, method, opts...)
				log.Ctx(ctx).Debug().Str("method", method).
					Dur("duration", time.Since(start)).
					Str("type", "stream").
					Err(err).
					Msg("GRPC requested")
				return cs, err
			},
		),
		grpc.WithChainUnaryInterceptor(
			clientMetrics.UnaryClientInterceptor(),
			grpc_sentry.UnaryClientInterceptor(),
			grpc_retry.UnaryClientInterceptor(opts...),
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				start := time.Now()
				err := invoker(prepareClientContext(ctx), method, req, reply, cc, opts...)
				log.Ctx(ctx).Debug().Str("method", method).
					Dur("duration", time.Since(start)).
					Str("type", "unary").
					Err(err).
					Msg("GRPC requested")
				return err
			},
		),
	)
	return conn, err
}

func prepareClientContext(ctx context.Context) context.Context {
	if loc, ok := locale.FromCtx(ctx); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, MetadataKeyLocale, loc.Serialize())
	}
	if token, ok := security.GetTokenFromContext(ctx); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, MetadataKeyBearerToken, token.GetValue())
	}
	if reqID := log.RequestIDFromContext(ctx); reqID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, MetadataKeyRequestID, reqID)
	}
	ctx = EncodeContextWithUTMData(ctx)

	if dep := middleware.ExternalDependencyContextFromContext(ctx); dep != nil {
		ctx = metadata.AppendToOutgoingContext(ctx, MetadataKeyExternalDependencies, dep.String())
	}

	return ctx
}
