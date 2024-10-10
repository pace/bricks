// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// RequestInContext stores a representation of the request in the
// context of said request. Some information of that request can then be
// accessed through the context using functions of this package.
func RequestInContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxReq := ctxRequest{
			RemoteAddr:    r.RemoteAddr,
			XForwardedFor: r.Header.Get("X-Forwarded-For"),
			UserAgent:     r.Header.Get("User-Agent"),
		}
		r = r.WithContext(contextWithRequest(r.Context(), &ctxReq))
		next.ServeHTTP(w, r)
	})
}

// ContextTransfer copies a request representation from one context to another.
func ContextTransfer(ctx, targetCtx context.Context) context.Context {
	if r := requestFromContext(ctx); r != nil {
		return contextWithRequest(targetCtx, r)
	}

	return targetCtx
}

type ctxRequest struct {
	RemoteAddr    string // requester IP:port
	XForwardedFor string // X-Forwarded-For header
	UserAgent     string // User-Agent header
}

func contextWithRequest(ctx context.Context, ctxReq *ctxRequest) context.Context {
	return context.WithValue(ctx, (*ctxRequest)(nil), ctxReq)
}

func requestFromContext(ctx context.Context) *ctxRequest {
	if v := ctx.Value((*ctxRequest)(nil)); v != nil {
		if request, ok := v.(*ctxRequest); ok {
			return request
		}
	}

	return nil
}

// GetXForwardedForHeaderFromContext returns the X-Forwarded-For header value
// that would express that you forwarded the request that is stored in the
// context.
//
// If the remote address of the request is 12.34.56.78:9999 then the value is
// that remote ip without the port. If the request already includes this header,
// the remote ip is appended to the value of that header. For example if the
// request on top of the remote ip also includes the header "X-Forwarded-For:
// 100.100.100.100" then the resulting value is "100.100.100.100, 12.34.56.78".
//
// Returns ErrNotFound if the context does not have a request. Returns
// ErrInvalidRequest if the request in the context is malformed, for example
// because it does not have a remote address, which should never happen.
func GetXForwardedForHeaderFromContext(ctx context.Context) (string, error) {
	ctxReq := requestFromContext(ctx)
	if ctxReq == nil {
		return "", fmt.Errorf("getting request from context: %w", ErrNotFound)
	}

	xForwardedFor := ctxReq.XForwardedFor

	ip, _, err := net.SplitHostPort(ctxReq.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf(
			"%w (from context): could not get ip from remote address: %w",
			ErrInvalidRequest, err)
	}

	if ip == "" {
		return "", fmt.Errorf(
			"%w (from context): could not get ip from remote address: %q",
			ErrInvalidRequest, ctxReq.RemoteAddr)
	}

	if xForwardedFor != "" {
		xForwardedFor += ", "
	}

	return xForwardedFor + ip, nil
}

// GetUserAgentFromContext returns the User-Agent header value from the request
// that is stored in the context. Returns ErrNotFound if the context does not
// have a request.
func GetUserAgentFromContext(ctx context.Context) (string, error) {
	ctxReq := requestFromContext(ctx)
	if ctxReq == nil {
		return "", fmt.Errorf("getting request from context: %w", ErrNotFound)
	}

	return ctxReq.UserAgent, nil
}
