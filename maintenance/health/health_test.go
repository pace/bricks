// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package health

import (
	"io/ioutil"
	"log"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	Handler().ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected /health to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "OK\n" {
		t.Errorf("Expected health to return OK, got: %q", string(data[:]))
	}
}
