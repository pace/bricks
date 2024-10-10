// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/maintenance/log"
)

func TestNewDefaultTransportChain(t *testing.T) {
	old := os.Getenv("HTTP_TRANSPORT_DUMP")

	defer func() {
		err := os.Setenv("HTTP_TRANSPORT_DUMP", old)
		require.NoError(t, err)
	}()

	err := os.Setenv("HTTP_TRANSPORT_DUMP", "request,response,body")
	require.NoError(t, err)

	t.Run("Finalizer not set explicitly", func(t *testing.T) {
		b := "Hello World"
		tr := NewDefaultTransportChain()
		retry := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			retry++
			if retry == 5 {
				w.WriteHeader(http.StatusOK)

				_, err := fmt.Fprint(w, b)
				require.NoError(t, err)

				return
			}

			w.WriteHeader(http.StatusBadGateway)

			_, err := fmt.Fprint(w, b)
			require.NoError(t, err)
		}))

		req := httptest.NewRequest(http.MethodGet, ts.URL, nil)
		req = req.WithContext(log.WithContext(context.Background()))

		resp, err := tr.RoundTrip(req)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

		ts.Close()

		assert.Equal(t, retry, 5)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Expected readable body, got error: %q", err.Error())
		}

		if ex, got := b, string(body); ex != got {
			t.Errorf("Expected body %q, got %q", ex, got)
		}
	})

	t.Run("Finalizer given", func(t *testing.T) {
		tr := &transportWithBody{body: "abc"}
		dt := NewDefaultTransportChain().Final(tr)

		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		req = req.WithContext(log.WithContext(context.Background()))

		resp, err := dt.RoundTrip(req)
		if err != nil {
			t.Fatalf("Expected err to be nil, got %#v", err)
		}

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Expected readable body, got error: %q", err.Error())
		}

		if ex, got := tr.body, string(body); ex != got {
			t.Errorf("Expected body %q, got %q", ex, got)
		}
	})
}

type transportWithBody struct {
	// returned response as string
	body string
}

func (t *transportWithBody) RoundTrip(req *http.Request) (*http.Response, error) {
	body := io.NopCloser(bytes.NewReader([]byte(t.body)))
	resp := &http.Response{Body: body, StatusCode: http.StatusOK}

	return resp, nil
}
