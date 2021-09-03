// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/09/03 by Vincent Landgraf

package grpc

import (
	"context"
	"time"

	"github.com/pace/bricks/maintenance/log"
	"google.golang.org/grpc"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

func Dial(address string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithChainStreamInterceptor(
			grpc_opentracing.StreamClientInterceptor(),
			grpc_prometheus.StreamClientInterceptor,
			grpc_retry.StreamClientInterceptor(opts...),
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				start := time.Now()
				cs, err := streamer(ctx, desc, cc, method, opts...)
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
			grpc_prometheus.UnaryClientInterceptor,
			grpc_retry.UnaryClientInterceptor(opts...),
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				start := time.Now()
				err := invoker(ctx, method, req, reply, cc, opts...)
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
