// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package http

import (
	"os"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	// Defaults
	os.Setenv("ADDR", "")
	os.Setenv("PORT", "")
	os.Setenv("MAX_HEADER_BYTES", "")
	os.Setenv("IDLE_TIMEOUT", "")
	os.Setenv("READ_TIMEOUT", "")
	os.Setenv("WRITE_TIMEOUT", "")
	parseConfig()
	s := Server(nil)
	cases := []struct {
		env              string
		expected, actual interface{}
	}{
		{"ADDR", ":3000", s.Addr},
		{"MAX_HEADER_BYTES", 1048576, s.MaxHeaderBytes},
		{"IDLE_TIMEOUT", time.Hour, s.IdleTimeout},
		{"READ_TIMEOUT", time.Minute, s.ReadTimeout},
		{"WRITE_TIMEOUT", time.Minute, s.WriteTimeout},
	}
	for _, tc := range cases {
		if tc.expected != tc.actual {
			t.Errorf("expected %s to be %v, got: %v", tc.env, tc.expected, tc.actual)
		}
	}

	// custom
	os.Setenv("ADDR", ":5432")
	os.Setenv("PORT", "1234")
	os.Setenv("MAX_HEADER_BYTES", "100")
	os.Setenv("IDLE_TIMEOUT", "1s")
	os.Setenv("READ_TIMEOUT", "2s")
	os.Setenv("WRITE_TIMEOUT", "3s")
	parseConfig()
	s = Server(nil)
	cases = []struct {
		env              string
		expected, actual interface{}
	}{
		{"ADDR", ":5432", s.Addr},
		{"MAX_HEADER_BYTES", 100, s.MaxHeaderBytes},
		{"IDLE_TIMEOUT", time.Second, s.IdleTimeout},
		{"READ_TIMEOUT", time.Second * 2, s.ReadTimeout},
		{"WRITE_TIMEOUT", time.Second * 3, s.WriteTimeout},
	}
	for _, tc := range cases {
		if tc.expected != tc.actual {
			t.Errorf("expected %s to be %v, got: %v", tc.env, tc.expected, tc.actual)
		}
	}
}

func TestEnvironment(t *testing.T) {
	// Defaults
	os.Setenv("ENVIRONMENT", "")
	parseConfig()
	if Environment() != "edge" {
		t.Errorf("Expected edge, got: %q", Environment())
	}

	// custom
	os.Setenv("ENVIRONMENT", "production")
	parseConfig()
	if Environment() != "production" {
		t.Errorf("Expected production, got: %q", Environment())
	}
}
