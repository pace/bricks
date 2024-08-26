package main

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pace/bricks/grpc"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/tools/testserver/math"
	"github.com/uber/jaeger-client-go"
)

type GrpcAuthBackend struct{}

func (*GrpcAuthBackend) AuthorizeStream(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (*GrpcAuthBackend) AuthorizeUnary(ctx context.Context) (context.Context, error) {
	token, ok := security.GetTokenFromContext(ctx)
	if ok {
		log.Ctx(ctx).Debug().Msgf("Token: %v", token.GetValue())
	} else {
		return nil, fmt.Errorf("unauthenticated")
	}
	return ctx, nil
}

type SimpleMathServer struct {
	math.UnimplementedMathServiceServer
}

func (*SimpleMathServer) Add(ctx context.Context, i *math.Input) (*math.Output, error) {
	if loc, ok := locale.FromCtx(ctx); ok {
		log.Ctx(ctx).Debug().Msgf("Locale: %q", loc.Serialize())
	}
	span := opentracing.SpanFromContext(ctx)
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		log.Ctx(ctx).Debug().Msgf("Span: %q", sc.String())
	}

	var o math.Output
	o.C = i.A + i.B
	log.Ctx(ctx).Debug().Msgf("A: %d + B: %d = C: %d", i.A, i.B, o.C)
	return &o, nil
}

func (*SimpleMathServer) Substract(ctx context.Context, i *math.Input) (*math.Output, error) {
	panic("not implemented")
}

func main() {
	ms := &SimpleMathServer{}
	gs := grpc.Server(&GrpcAuthBackend{})
	math.RegisterMathServiceServer(gs, ms)

	err := grpc.ListenAndServe(gs)
	if err != nil {
		log.Fatal(err)
	}
}
