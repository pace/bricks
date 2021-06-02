// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/05/10 by Alessandro Miceli

package middleware

import (
	"context"
	"net/http"
	"sync"

	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/maintenance/log"
)

// ClientIDHeaderName name of the HTTP header that is used for reporting
const ClientIDHeaderName = "Jwt-Client-ID"

// ResponseClientID to report jwt client id
func ResponseClientID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rcID ResponseClientIDContext
		rcw := responseClientIDWriter{
			ResponseWriter: w,
			rcID:           &rcID,
		}
		r = r.WithContext(ContextWithResponseClientID(r.Context(), &rcID))
		next.ServeHTTP(&rcw, r)
	})
}

func AddClientIDToResponse(ctx context.Context) {
	clientID, _ := oauth2.ClientID(ctx)
	rcID := ClientIDContextFromContext(ctx)
	if rcID == nil {
		log.Ctx(ctx).Warn().Msgf("can't add request client ID dependency %s, because context is missing", clientID)
		return
	}

	rcID.AddClientID(clientID)
}

type responseClientIDWriter struct {
	http.ResponseWriter
	rcID *ResponseClientIDContext
}

// ContextWithResponseClientID creates a contex with the provided dependencies
func ContextWithResponseClientID(ctx context.Context, rcID *ResponseClientIDContext) context.Context {
	return context.WithValue(ctx, (*ResponseClientIDContext)(nil), rcID)
}

// ClientIDContextFromContext returns the jwt-client-id dependency context or nil
func ClientIDContextFromContext(ctx context.Context) *ResponseClientIDContext {
	if v := ctx.Value((*ResponseClientIDContext)(nil)); v != nil {
		return v.(*ResponseClientIDContext)
	}
	return nil
}

// ResponseClientIDContext contains the clientID
type ResponseClientIDContext struct {
	mu       sync.RWMutex
	clientID string
}

func (rcID *ResponseClientIDContext) AddClientID(clientID string) {
	rcID.mu.Lock()
	rcID.clientID = clientID
	rcID.mu.Unlock()
}
