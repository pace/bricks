// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pace/bricks/grpc"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/locale"

	"github.com/pace/bricks/maintenance/failover"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"

	"github.com/pace/bricks/backend/couchdb"
	"github.com/pace/bricks/backend/objstore"
	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/backend/redis"
	pacehttp "github.com/pace/bricks/http"
	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	_ "github.com/pace/bricks/maintenance/tracing"
	"github.com/pace/bricks/test/livetest"
	"github.com/pace/bricks/tools/testserver/math"
	simple "github.com/pace/bricks/tools/testserver/simple"
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

type TestService struct{}

func (*TestService) GetTest(ctx context.Context, w simple.GetTestResponseWriter, r *simple.GetTestRequest) error {
	log.Debug("Request in flight, this will wait 5 min....")
	for t := 0; t < 360; t++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Second)
		}
	}
	return nil
}

func main() {
	db := postgres.NewDB(context.Background())
	rdb := redis.Client()
	cdb, err := couchdb.DefaultDatabase()
	if err != nil {
		log.Fatal(err)
	}
	_, err = objstore.Client()
	if err != nil {
		log.Fatal(err)
	}

	ap, err := failover.NewActivePassive("testserver", time.Second*10, rdb)
	if err != nil {
		log.Fatal(err)
	}
	go ap.Run(log.WithContext(context.Background())) // nolint: errcheck

	h := pacehttp.Router()
	servicehealthcheck.RegisterHealthCheckFunc("fail-50", func(ctx context.Context) (r servicehealthcheck.HealthCheckResult) {
		if time.Now().Unix()%2 == 0 {
			panic("boom")
		}
		r.Msg = "Foo"
		r.State = servicehealthcheck.Ok
		return
	})

	h.Handle("/pay/beta/test", simple.Router(new(TestService)))

	h.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		span := sentry.StartSpan(ctx, "TestHandler")
		defer span.Finish()

		ctx = span.Context()

		// do dummy database query
		var result struct {
			Calc int //nolint
		}
		res, err := db.NewSelect().Model(&result).ColumnExpr("? + ? AS Calc", 10, 10).Exec(ctx)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Calc failed")
			return
		}

		count, err := res.RowsAffected()
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("RowsAffected failed")
			return
		}

		log.Ctx(ctx).Debug().Int64("rows_affected", count).Msg("Calc done")

		// do dummy redis query
		crdb := redis.WithContext(ctx, rdb)
		if err := crdb.Ping(ctx).Err(); err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Ping failed")
			return
		}

		// do dummy call to external service
		log.Ctx(ctx).Debug().Msg("Test before JSON")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"street":"Haid-und-Neu-Straße 18, 76131 Karlsruhe", "sunset": "%s"}`, fetchSunsetandSunrise(ctx))
	})

	h.HandleFunc("/grpc", func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		conn, err := grpc.NewClient(":3001")
		if err != nil {
			log.Fatalf("did not connect: %s", err)
		}
		defer conn.Close()

		ctx = security.ContextWithToken(ctx, security.TokenString("test"))

		c := math.NewMathServiceClient(conn)
		o, err := c.Add(ctx, &math.Input{
			A: 1,
			B: 23,
		})
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("failed to add")
			return
		}
		log.Ctx(ctx).Info().Msgf("C: %d", o.C)

		ctx = locale.WithLocale(ctx, locale.NewLocale("fr-CH", "Europe/Paris"))

		_, err = c.Add(ctx, &math.Input{})
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("failed to substract")
			return
		}

		if r.URL.Query().Get("error") != "" {
			_, err = c.Substract(ctx, &math.Input{})
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("failed to substract")
				return
			}
		}
	})

	h.HandleFunc("/couch", func(w http.ResponseWriter, r *http.Request) {
		row := cdb.Get(r.Context(), "$health_check")
		if err := row.Err(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var doc interface{}
		row.ScanDoc(&doc) // nolint: errcheck

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc) // nolint: errcheck
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
	m := oauth2.NewMiddleware(new(OauthBackend)) // nolint: staticcheck

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
				return
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
	span := sentry.StartSpan(ctx, "fetchSunsetandSunrise")
	defer span.Finish()

	ctx = span.Context()

	span.SetData("lat", lat)
	span.SetData("lon", lon)

	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f&date=today", lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	c := &http.Client{
		Transport: transport.NewDefaultTransportChain(),
	}
	resp, err := c.Do(req)
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
