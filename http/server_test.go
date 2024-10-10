// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package http

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	// Defaults
	err := os.Setenv("ADDR", "")
	require.NoError(t, err)

	err = os.Setenv("PORT", "")
	require.NoError(t, err)

	err = os.Setenv("MAX_HEADER_BYTES", "")
	require.NoError(t, err)

	err = os.Setenv("IDLE_TIMEOUT", "")
	require.NoError(t, err)

	err = os.Setenv("READ_TIMEOUT", "")
	require.NoError(t, err)

	err = os.Setenv("WRITE_TIMEOUT", "")
	require.NoError(t, err)

	parseConfig()

	s := Server(nil)

	cases := []struct {
		env              string
		expected, actual any
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
	err = os.Setenv("ADDR", ":5432")
	require.NoError(t, err)

	err = os.Setenv("PORT", "1234")
	require.NoError(t, err)

	err = os.Setenv("MAX_HEADER_BYTES", "100")
	require.NoError(t, err)

	err = os.Setenv("IDLE_TIMEOUT", "1s")
	require.NoError(t, err)

	err = os.Setenv("READ_TIMEOUT", "2s")
	require.NoError(t, err)

	err = os.Setenv("WRITE_TIMEOUT", "3s")
	require.NoError(t, err)

	parseConfig()

	s = Server(nil)

	cases = []struct {
		env              string
		expected, actual any
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
	err := os.Setenv("ENVIRONMENT", "")
	require.NoError(t, err)

	parseConfig()

	if Environment() != "edge" {
		t.Errorf("Expected edge, got: %q", Environment())
	}

	// custom
	err = os.Setenv("ENVIRONMENT", "production")
	require.NoError(t, err)

	parseConfig()

	if Environment() != "production" {
		t.Errorf("Expected production, got: %q", Environment())
	}
}
