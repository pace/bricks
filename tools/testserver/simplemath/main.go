package main

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"

	"github.com/pace/bricks/grpc"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/tools/testserver/math"
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

	span := sentry.SpanFromContext(ctx)
	if span != nil {
		log.Ctx(ctx).Debug().Msgf("Span: %q", span.Name)
	}

	var o math.Output

	o.C = i.GetA() + i.GetB()
	log.Ctx(ctx).Debug().Msgf("A: %d + B: %d = C: %d", i.GetA(), i.GetB(), o.GetC())

	return &o, nil
}

func (*SimpleMathServer) Subtract(ctx context.Context, i *math.Input) (*math.Output, error) {
	panic("not implemented")
}

func main() {
	l := log.Logger()
	ms := &SimpleMathServer{}
	gs := grpc.Server(&GrpcAuthBackend{}, log.InterceptorLogger(l))
	math.RegisterMathServiceServer(gs, ms)

	if err := grpc.ListenAndServe(gs); err != nil {
		log.Fatal(err)
	}
}
