// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/09/03 by Vincent Landgraf

package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/log"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
)

func DialContext(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return dialCtx(ctx, addr)
}

func Dial(addr string) (*grpc.ClientConn, error) {
	return dialCtx(context.Background(), addr)
}

func dialCtx(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn

	clientMetrics := grpc_prometheus.NewClientMetrics()

	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
	}
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(
			grpc_opentracing.StreamClientInterceptor(),
			grpc_opentracing.StreamClientInterceptor(),
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
			grpc_opentracing.UnaryClientInterceptor(),
			clientMetrics.UnaryClientInterceptor(),
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
		ctx = metadata.AppendToOutgoingContext(ctx, "locale", loc.Serialize())
	}
	if token, ok := security.GetTokenFromContext(ctx); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, "bearer_token", token.GetValue())
	}
	if reqID := log.RequestIDFromContext(ctx); reqID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "req_id", reqID)
	}
	ctx = EncodeContextWithUTMData(ctx)
	return ctx
}
