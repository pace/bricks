// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

package oauth2

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Example() {
	r := mux.NewRouter()
	middleware := Middleware{
		URL:          "http://localhost:3000",
		ClientID:     "13972c02189a6e938a4730bc81c2a20cc4e03ef5406d20d2150110584d6b3e6c",
		ClientSecret: "7d26f8918a83bd155a936bbe780f32503a88cb8bd3e8acf25248357dff31668e",
	}

	r.Use(middleware.Handler)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AUDIT: User %s does something", UserID(r.Context()))

		if HasScope(r.Context(), "dtc:codes:write") {
			fmt.Fprintf(w, "User has scope.")
			return
		}

		fmt.Fprintf(w, "Your client may not have the right scopes to see the secret code")
	})

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
	}

	log.Fatal(srv.ListenAndServe())
}
