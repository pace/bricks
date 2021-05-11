// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/05/10 by Alessandro Miceli

package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/pace/bricks/http/oauth2"
)

// ClientIDHeaderName name of the HTTP header that is used for reporting
const ClientIDHeaderName = "Client-ID"

// ResponseClientID middleware to report client ID
func ResponseClientID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rcc ResponseClientIDContext
		rcw := responseClientIDWriter{
			ResponseWriter: w,
			rcc:            &rcc,
		}

		r = r.WithContext(ContextWithResponseClientID(r.Context(), &rcc))
		next.ServeHTTP(&rcw, r)
	})
}

func AddResponseClientID(ctx context.Context) {
	clientID, _ := oauth2.ClientID(ctx)
	cIDc := ResponseClientIDContextFromContext(ctx)
	if cIDc == nil {
		//log.Ctx(ctx).Warn().Msgf("can't add client %s, because context is missing", clientID)
		cIDc = &ResponseClientIDContext{}
		//return
	}
	cIDc.AddResponseClientID(clientID)
}

type responseClientIDWriter struct {
	http.ResponseWriter
	header bool
	rcc    *ResponseClientIDContext
}

// addHeader adds the clientID header if not done already
func (w *responseClientIDWriter) addHeader() {
	if len(w.rcc.clientIDs) > 0 {
		w.ResponseWriter.Header().Add(ClientIDHeaderName, w.rcc.String())
	}
	w.header = true
}

func (w *responseClientIDWriter) Write(data []byte) (int, error) {
	w.addHeader()
	return w.ResponseWriter.Write(data)
}

// ContextWithResponseClientID creates a contex with the provided client ID
func ContextWithResponseClientID(ctx context.Context, rcc *ResponseClientIDContext) context.Context {
	return context.WithValue(ctx, (*ResponseClientIDContext)(nil), rcc)
}

// ResponseClientIDContextFromContext returns the client ID context or nil
func ResponseClientIDContextFromContext(ctx context.Context) *ResponseClientIDContext {
	if v := ctx.Value((*ResponseClientIDContext)(nil)); v != nil {
		return v.(*ResponseClientIDContext)
	}
	return nil
}

// ResponseClientIDContext contains all client IDs that were seen
// during the request livecycle
type ResponseClientIDContext struct {
	mu        sync.RWMutex
	clientIDs []responseClientID
}

func (rcc *ResponseClientIDContext) AddResponseClientID(clientID string) {
	rcc.mu.Lock()
	rcc.clientIDs = append(rcc.clientIDs, responseClientID{
		ClientID: clientID,
	})
	rcc.mu.Unlock()
}

// String formats all client IDs
func (rcc *ResponseClientIDContext) String() string {
	var b strings.Builder
	sep := len(rcc.clientIDs) - 1
	for _, dep := range rcc.clientIDs {
		b.WriteString(dep.String())
		if sep > 0 {
			b.WriteByte(',')
			sep--
		}
	}
	return b.String()
}

// responseClientID represents the client ID that
// was sent in the request
type responseClientID struct {
	ClientID string //client ID name
}

// String return a client ID
func (rcc responseClientID) String() string {
	return rcc.ClientID
}
