// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	pacehttp "lab.jamit.de/pace/go-microservice/http"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
	_ "lab.jamit.de/pace/go-microservice/maintenance/tracing"
)

func main() {
	h := pacehttp.Router()
	h.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("TestHandler", opentracing.ChildOf(wireContext))
		defer handlerSpan.Finish()
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)

		log.Ctx(ctx).Debug().Msg("Test before JSON")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"hello":"world", "time": "%s"}`, getTime(ctx))
	})
	s := pacehttp.Server(h)
	log.Logger().Info().Str("addr", s.Addr).Msg("Starting testserver ...")
	log.Fatal(s.ListenAndServe())
}

func getTime(ctx context.Context) string {
	t := time.Now()
	log.Ctx(ctx).Debug().Time("gentime", t).Msg("generating time")
	return t.String()
}
