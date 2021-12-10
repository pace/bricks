package utm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	m := Middleware()
	req := httptest.NewRequest(http.MethodGet, "http://example.org/?utm_source=internet", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, found := FromContext(r.Context())
		assert.True(t, found)
		assert.Equal(t, UTMData{
			Source: "internet",
		}, data)
	})
	m(h).ServeHTTP(w, req)
}

func TestRoundTripper_RoundTrip(t *testing.T) {
	tripper := &RoundTripper{
		transport: &mockTripper{
			t:    t,
			resp: &http.Response{},
			requiredQueryParameters: map[string]string{
				"utm_source":         "src",
				"utm_medium":         "mdm",
				"utm_campaign":       "camp",
				"utm_term":           "trm",
				"utm_content":        "cnts",
				"utm_partner_client": "clnt",
			},
		},
	}
	ctx := ContextWithUTMData(context.Background(), UTMData{
		Source:   "src",
		Medium:   "mdm",
		Campaign: "camp",
		Term:     "trm",
		Content:  "cnts",
		Client:   "clnt",
	})
	req := httptest.NewRequest(http.MethodGet, "http://example.org/?utm_source=internet", nil)
	req = req.WithContext(ctx)
	resp, err := tripper.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

var _ http.RoundTripper = (*mockTripper)(nil)

type mockTripper struct {
	t                       *testing.T
	resp                    *http.Response
	requiredQueryParameters map[string]string
}

func (m *mockTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.requiredQueryParameters != nil {
		for k, v := range m.requiredQueryParameters {
			h := req.URL.Query().Get(k)
			assert.Equal(m.t, v, h, fmt.Sprintf("expected query paramater %q to match value", k))
		}
	}
	return m.resp, nil
}
