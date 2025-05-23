// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.

package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"
	"github.com/pace/bricks/http/middleware"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/log/hlog"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/caarlos0/env/v11"
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

func Server(ab AuthBackend, logger grpc_logging.Logger) *grpc.Server {
	serverMetrics := grpc_prometheus.NewServerMetrics()

	myServer := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			grpc_sentry.StreamServerInterceptor(),
			grpc_logging.StreamServerInterceptor(logger),
			serverMetrics.StreamServerInterceptor(),
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

				if err := addExternalDependencyToTrailer(ctx); err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("unable to add external dependencies to trailer")
				}

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
			func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				defer errors.HandleWithCtx(stream.Context(), "GRPC "+info.FullMethod)
				err = InternalServerError // default in case of a panic
				err = handler(srv, stream)
				return err
			},
			grpc_auth.StreamServerInterceptor(ab.AuthorizeStream),
		),
		grpc.ChainUnaryInterceptor(
			grpc_sentry.UnaryServerInterceptor(),
			grpc_logging.UnaryServerInterceptor(logger),
			serverMetrics.UnaryServerInterceptor(),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctx, md := prepareContext(ctx)

				var addr string
				if p, ok := peer.FromContext(ctx); ok {
					addr = p.Addr.String()
				}

				start := time.Now()
				resp, err = handler(ctx, req)

				if err := addExternalDependencyToTrailer(ctx); err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("unable to add external dependencies to trailer")
				}

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
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				defer errors.HandleWithCtx(ctx, "GRPC "+info.FullMethod)
				err = InternalServerError // default in case of a panic
				resp, err = handler(ctx, req)
				return
			},
			grpc_auth.UnaryServerInterceptor(ab.AuthorizeUnary),
		),
	)

	return myServer
}

// addExternalDependencyToTrailer adds the external dependencies to the grpc trailer.
// This is used to track external dependencies which are updated during the request lifecycle.
func addExternalDependencyToTrailer(ctx context.Context) error {
	edc := middleware.ExternalDependencyContextFromContext(ctx)
	if edc == nil {
		return nil
	}

	md := metadata.Pairs(MetadataKeyExternalDependencies, edc.String())

	return grpc.SetTrailer(ctx, md)
}

func prepareContext(ctx context.Context) (context.Context, metadata.MD) {
	md, _ := metadata.FromIncomingContext(ctx)
	logger := zlog.With().Logger()

	// add request context if req_id is given
	var reqID xid.ID
	if ri := md.Get(MetadataKeyRequestID); len(ri) > 0 {
		var err error
		reqID, err = xid.FromString(ri[0])
		if err != nil {
			log.Debugf("unable to parse xid from req_id: %v", err)
			reqID = xid.New()
		}
	} else {
		// generate random request id
		reqID = xid.New()
	}

	//  attach request ID to context and logger
	ctx = hlog.WithValue(ctx, reqID)

	// set logger and log sink
	ctx = log.ContextWithSink(logger.WithContext(ctx), log.NewSink())
	zlog := zerolog.Ctx(ctx)
	zlog.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("req_id", reqID.String())
	})

	// handle locale
	if l := md.Get(MetadataKeyLocale); len(l) > 0 {
		loc, err := locale.ParseLocale(l[0])
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msgf("unable to parse locale: %v", err)
		} else {
			ctx = locale.WithLocale(ctx, loc)
		}
	}

	ctx = ContextWithUTMFromMetadata(ctx, md)

	// add security context if bearer token is given
	if bt := md.Get(MetadataKeyBearerToken); len(bt) > 0 {
		ctx = security.ContextWithToken(ctx, security.TokenString(bt[0]))
	}

	// add external dependencies to context
	externalDependencyContext := middleware.ExternalDependencyContext{}

	if externalDependencies := md.Get(MetadataKeyExternalDependencies); len(externalDependencies) > 0 {
		externalDependencyContext.Parse(externalDependencies[0])
	}

	ctx = middleware.ContextWithExternalDependency(ctx, &externalDependencyContext)

	delete(md, MetadataKeyContentType)
	delete(md, MetadataKeyLocale)
	delete(md, MetadataKeyBearerToken)
	delete(md, MetadataKeyRequestID)
	delete(md, MetadataKeyExternalDependencies)

	return ctx, md
}
