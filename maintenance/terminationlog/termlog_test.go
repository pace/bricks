package terminationlog

import (
	"errors"
	"fmt"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestExtractErrorFrames(t *testing.T) {

	tcs := []struct {
		name       string
		err        error
		shouldFail bool
	}{
		{
			name:       "error created with std lib",
			err:        errors.New("error"),
			shouldFail: true,
		},
		{
			name:       "error created with github pkg",
			err:        pkgerrors.New("error"),
			shouldFail: false,
		},
		{
			name:       "plain error",
			err:        fmt.Errorf("error"),
			shouldFail: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := extractErrorFrames(tc.err)
			assert.Equal(t, err != nil, tc.shouldFail)
		})
	}
}

func TestBuildTerminationLogOutput(t *testing.T) {
	errMsg := "error"

	tcs := []struct {
		name                 string
		errFunc              func(s string) error
		outputWithStackTrace bool
	}{
		{
			name: "error created with github pkg",
			errFunc: func(s string) error {
				return pkgerrors.New(s)
			},
			outputWithStackTrace: true,
		},
		{
			name: "error created with std lib",
			errFunc: func(s string) error {
				return errors.New(s)
			},
			outputWithStackTrace: false,
		},
		{
			name: "plain error",
			errFunc: func(s string) error {
				return fmt.Errorf(s)
			},
			outputWithStackTrace: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := buildTerminationLogOutput(tc.errFunc(errMsg))
			assert.Equal(t, tc.outputWithStackTrace, len(out) > len(fmt.Sprintf("%s\n", errMsg)))
			out = buildTerminationLogOutputf("message: %v", tc.errFunc(errMsg))
			assert.Equal(t, tc.outputWithStackTrace, len(out) > len(fmt.Sprintf("message: %s\n", errMsg)))
		})
	}
}
