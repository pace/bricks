// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/09/03 by Vincent Landgraf

package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/log/hlog"
	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/caarlos0/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var InternalServerError = errors.New("internal server error")

type Config struct {
	Address string `env:"GRPC_ADDR" envDefault:":3001"`
}

type AuthBackend interface {
	AuthorizeStream(ctx context.Context) (context.Context, error)
	AuthorizeUnary(ctx context.Context) (context.Context, error)
}

func ListenAndServe(gs *grpc.Server) error {
	listener, err := Listener()
	if err != nil {
		return err
	}
	log.Logger().Info().Str("addr", listener.Addr().String()).Msg("Starting grpc server ...")
	err = gs.Serve(listener)
	if err != nil {
		return err
	}
	return nil
}

func Listener() (net.Listener, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grpc server environment: %w", err)
	}

	tcpListener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to create grpc listener for %q: %w", cfg.Address, err)
	}
	return tcpListener, nil
}

func Server(ab AuthBackend) *grpc.Server {
	myServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				ctx := stream.Context()
				ctx, md := prepareContext(ctx)

				wrappedStream := grpc_middleware.WrapServerStream(stream)
				wrappedStream.WrappedContext = ctx
				var addr string
				if p, ok := peer.FromContext(ctx); ok {
					addr = p.Addr.String()
				}

				start := time.Now()
				err := handler(srv, wrappedStream)

				log.Ctx(ctx).Info().Str("method", info.FullMethod).
					Dur("duration", time.Since(start)).
					Str("type", "stream").
					Str("ip", addr).
					Interface("md", md).
					Str("user_agent", strings.Join(md.Get("user-agent"), ",")).
					Err(err).
					Msg("GRPC completed Stream")
				return err
			},
			grpc_auth.StreamServerInterceptor(ab.AuthorizeStream),
			func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				defer errors.HandleWithCtx(stream.Context(), "GRPC "+info.FullMethod)
				err = InternalServerError // default in case of a panic
				err = handler(srv, stream)
				return err
			},
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctx, md := prepareContext(ctx)

				var addr string
				if p, ok := peer.FromContext(ctx); ok {
					addr = p.Addr.String()
				}

				start := time.Now()
				resp, err = handler(ctx, req)

				log.Ctx(ctx).Info().Str("method", info.FullMethod).
					Dur("duration", time.Since(start)).
					Str("type", "unary").
					Str("ip", addr).
					Interface("md", md).
					Str("user_agent", strings.Join(md.Get("user-agent"), ",")).
					Err(err).
					Msg("GRPC completed Unary")
				return
			},
			grpc_auth.UnaryServerInterceptor(ab.AuthorizeUnary),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				defer errors.HandleWithCtx(ctx, "GRPC "+info.FullMethod)
				err = InternalServerError // default in case of a panic
				resp, err = handler(ctx, req)
				return
			},
		)),
	)

	return myServer
}

func prepareContext(ctx context.Context) (context.Context, metadata.MD) {
	md, _ := metadata.FromIncomingContext(ctx)
	delete(md, "content-type")

	logger := log.Logger().With().Logger()

	// add request context if req_id is given
	var reqID xid.ID
	if ri := md.Get("req_id"); len(ri) > 0 {
		var err error
		reqID, err = xid.FromString(ri[0])
		if err != nil {
			log.Debugf("unable to parse xid from req_id: %v", err)
			reqID = xid.New()
		}
		delete(md, "req_id")
	} else {
		// generate random request id
		reqID = xid.New()
	}

	// handle locale
	if l := md.Get("locale"); len(l) > 0 {
		loc, err := locale.ParseLocale(l[0])
		if err != nil {
			log.Debugf("unable to parse locale: %v", err)
		} else {
			ctx = locale.WithLocale(ctx, loc)
		}
		delete(md, "locale")
	}

	//  attach request ID to context and logger
	ctx = hlog.WithValue(ctx, reqID)
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("req_id", reqID.String())
	})

	// set logger and log sink
	ctx = log.ContextWithSink(logger.WithContext(ctx), log.NewSink())

	// add security context if bearer token is given
	if bt := md.Get("bearer_token"); len(bt) > 0 {
		token := bt[0]
		// cut bearer token to not show up in logs
		if index := strings.LastIndex(token, "."); index != -1 {
			// cut jwt signature
			md.Set("bearer_token", token[:index])
		} else {
			// otherwise show half of the token stared out
			md.Set("bearer_token", token[:len(token)/2]+strings.Repeat("*", len(token)/2))
		}
		ctx = security.ContextWithToken(ctx, security.TokenString(token))
	}

	return ctx, md
}
