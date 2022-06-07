package terminationlog

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractErrorFrames(t *testing.T) {

	tcs := []struct {
		name       string
		err        error
		shouldFail bool
	}{
		{
			name:       "error created with github pkg",
			err:        errors.New("error"),
			shouldFail: false,
		},
		{
			name:       "error created with std lib",
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
