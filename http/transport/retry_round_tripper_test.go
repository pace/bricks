// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetryRoundTripper(t *testing.T) {
	type args struct {
		requestBody []byte
		statuses    []int
		err         error
	}

	anyErr := errors.New("any error")

	tests := []struct {
		name        string
		args        args
		wantRetries int
		wantErr     error
	}{
		{
			name: "Successful response after some retries",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				statuses:    []int{408, 502, 503, 504, 200},
			},
			wantRetries: 5,
		},
		{
			name: "No retry after error response",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				err:         anyErr,
			},
			wantRetries: 0,
			wantErr:     anyErr,
		},
		{
			name: "No retry after context is canceled",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				err:         context.Canceled,
			},
			wantRetries: 0,
			wantErr:     context.Canceled,
		},
		{
			name: "No retry after context deadline is exceeded",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				err:         context.DeadlineExceeded,
			},
			wantRetries: 0,
			wantErr:     context.DeadlineExceeded,
		},
		{
			name: "Exceed retries",
			args: args{
				requestBody: []byte(`{"key":"value""}`),
				statuses:    []int{408, 502, 503, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 504, 200},
			},
			wantRetries: 10,
			wantErr:     ErrRetryFailed,
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

			req := httptest.NewRequest("GET", "/foo", bytes.NewReader(tt.args.requestBody))
			resp, err := rt.RoundTrip(req.WithContext(context.Background()))

			require.Equal(t, tt.wantRetries, tr.attempts)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
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
	t.ctx = req.Context()

	if t.err != nil {
		return nil, fmt.Errorf("%w", t.err)
	}
	readAll, _ := io.ReadAll(req.Body)
	body := io.NopCloser(bytes.NewReader(readAll))
	resp := &http.Response{Body: body, StatusCode: t.statusCodes[t.attempts]}
	t.attempts++

	return resp, nil
}
