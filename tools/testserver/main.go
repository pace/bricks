// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/backend/redis"
	pacehttp "github.com/pace/bricks/http"
	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	_ "github.com/pace/bricks/maintenance/tracing"
	"github.com/pace/bricks/test/livetest"
)

// pace lat/lon
var (
	lat = 49.012553
	lon = 8.427087
)

type OauthBackend struct{}

func (*OauthBackend) IntrospectToken(ctx context.Context, token string) (*oauth2.IntrospectResponse, error) {
	return &oauth2.IntrospectResponse{
		Active:   true,
		ClientID: "some client",
		Scope:    "email profile",
		UserID:   "285ec1fc-2843-4ed8-bfa8-4217880c8348",
	}, nil
}

func main() {
	db := postgres.DefaultConnectionPool()
	rdb := redis.Client()

	h := pacehttp.Router()

	h.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "TestHandler")
		defer handlerSpan.Finish()

		// do dummy database query
		cdb := db.WithContext(ctx)
		var result struct {
			Calc int //nolint
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

	// Test OAuth
	//
	// This middleware is configured against an Oauth application dummy
	m := oauth2.Middleware{Backend: new(OauthBackend)}

	sr := h.PathPrefix("/test").Subrouter()
	sr.Use(m.Handler)

	// To actually test the Oauth2, one can run the following as an example:
	//
	// curl -H "Authorization: Bearer 83142f1b767e910e78ba2d554b6708c371f053d13d6075bcc39766853a932253" localhost:3000/test/auth
	sr.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Oauth test successful.\n")
	})

	s := pacehttp.Server(h)
	log.Logger().Info().Str("addr", s.Addr).Msg("Starting testserver ...")

	// nolint:errcheck
	go livetest.Test(context.Background(), []livetest.TestFunc{
		func(t *livetest.T) {
			t.Log("Test /test query")

			resp, err := http.Get("http://localhost:3000/test")
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			if resp.StatusCode != 200 {
				t.Logf("Received status code: %d", resp.StatusCode)
				t.Fail()
			}
		},
	})

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
