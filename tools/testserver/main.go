// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pacehttp "lab.jamit.de/pace/go-microservice/http"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

func main() {
	h := pacehttp.Router()
	h.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Req(r).Debug().Msg("Test before JSON")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"hello":"world", "time": "%s"}`, getTime(r.Context()))
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
