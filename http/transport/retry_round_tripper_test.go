// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/15 by Florian Hübsch

package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetryRoundTripper(t *testing.T) {

	type args struct {
		requestBody []byte
		statuses    []int
		err         error
		cancel      bool
	}

	tests := []struct {
		name        string
		args        args
		wantRetries int
		wantErr     bool
	}{
		{
			name: "Successful response after some retries",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				statuses:    []int{408, 502, 503, 504, 200},
			},
			wantRetries: 5,
			wantErr:     false,
		},
		{
			name: "No retry after error response",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				statuses:    []int{408, 502, 503, 504, 200},
				err:         errors.New("Any Error"),
			},
			wantRetries: 0,
			wantErr:     true,
		},
		{
			name: "No retry after context is finished",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				statuses:    []int{408, 502, 503, 504, 200},
				cancel:      true,
			},
			wantRetries: 1,
			wantErr:     true,
		},
		{
			name: "Exceed retries",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				statuses:    []int{408, 502, 503, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 200},
			},
			wantRetries: 10,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rt := NewDefaultRetryRoundTripper()
			tr := &retriedTransport{
				statusCodes: tt.args.statuses,
				err:         tt.args.err,
			}
			rt.SetTransport(tr)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// cancel directly, so the original request is performed and
			// then the retry mechanism aborts on the second attempt
			if tt.args.cancel {
				cancel()
			}
			req := httptest.NewRequest("GET", "/foo", bytes.NewReader(tt.args.requestBody))
			resp, err := rt.RoundTrip(req.WithContext(ctx))

			require.Equal(t, tt.wantRetries, tr.attempts)

			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, string(tt.args.requestBody), string(body))
			require.Equal(t, tt.wantRetries, int(attemptFromCtx(tr.ctx)))

		})
	}
}

type retriedTransport struct {
	// number of attempts
	attempts int
	// returned status codes in order they are provided
	statusCodes []int
	// returned error
	err error
	// recorded context
	ctx context.Context
}

func (t *retriedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("Request Nr %d %d\n", t.attempts, t.statusCodes[t.attempts])
	t.ctx = req.Context()

	if t.err != nil {
		return nil, t.err
	}
	readAll, _ := io.ReadAll(req.Body)
	body := ioutil.NopCloser(bytes.NewReader(readAll))
	resp := &http.Response{Body: body, StatusCode: t.statusCodes[t.attempts]}
	t.attempts++

	return resp, nil
}
