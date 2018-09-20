// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"lab.jamit.de/pace/go-microservice/backend/postgres"
	"lab.jamit.de/pace/go-microservice/backend/redis"
	pacehttp "lab.jamit.de/pace/go-microservice/http"
	"lab.jamit.de/pace/go-microservice/maintenance/errors"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
	_ "lab.jamit.de/pace/go-microservice/maintenance/tracing"
)

// pace lat/lon
var (
	lat = 49.012553
	lon = 8.427087
)

func main() {
	db := postgres.ConnectionPool()
	rdb := redis.Client()
	h := pacehttp.Router()

	h.Use(errors.Handler())
	h.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// add opentracing span + context
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("TestHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		defer handlerSpan.Finish()

		// do dummy database query
		cdb := db.WithContext(ctx)
		var result struct {
			Calc int
		}
		res, err := cdb.QueryOne(&result, `SELECT ? + ? AS Calc`, 10, 10)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Calc failed")
			return
		}
		log.Ctx(ctx).Debug().Int("rows_affected", res.RowsAffected()).Msg("Calc done")

		// do dummy redis query
		crdb := redis.WithContext(ctx, rdb)
		if err := crdb.Ping().Err(); err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Ping failed")
			return
		}

		// do dummy call to external service
		log.Ctx(ctx).Debug().Msg("Test before JSON")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"street":"Haid-und-Neu-Straße 18, 76131 Karlsruhe", "sunset": "%s"}`, fetchSunsetandSunrise(ctx))
	})
	h.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			defer errors.HandleWithCtx(r.Context(), "Some worker")

			panic(fmt.Errorf("Something went wrong %d - times", 100))
		}()

		panic("Test for sentry")
	})
	h.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		errors.HandleError(errors.WrapWithExtra(errors.New("Wrap error"), map[string]interface{}{
			"Foo": 123,
		}), "wrapHandler", w, r)
	})
	s := pacehttp.Server(h)
	log.Logger().Info().Str("addr", s.Addr).Msg("Starting testserver ...")
	log.Fatal(s.ListenAndServe())
}

func fetchSunsetandSunrise(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "fetchSunsetandSunrise")
	defer span.Finish()
	span.LogFields(olog.Float64("lat", lat), olog.Float64("lon", lon))

	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f&date=today", lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var r struct {
		Results struct {
			Sunset string `json:"sunset"`
		} `json:"results"`
	}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		log.Fatal(err)
	}

	sunset, err := time.Parse("3:04:05 PM", r.Results.Sunset)
	if err != nil {
		log.Fatal(err)
	}
	sunset = sunset.Local()

	log.Ctx(ctx).Debug().Time("sunset", sunset).Str("str", r.Results.Sunset).Msg("Parsed sunset time")
	return sunset.String()
}
