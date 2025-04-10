package objstore

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	http2 "github.com/pace/bricks/http"
	"github.com/pace/bricks/maintenance/log"
)

func setup() *http.Response {
	r := http2.Router()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/check", nil)
	r.ServeHTTP(rec, req)

	return rec.Result()
}

// TestIntegrationHealthCheck tests if object storage health check ist working like expected.
func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	RegisterHealthchecks()
	time.Sleep(1 * time.Second) // by the magic of asynchronous code, I here-by present a magic wait

	resp := setup()

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected /health/check to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if !strings.Contains(string(data), "objstore               OK") {
		t.Errorf("Expected /health/check to return OK, got: %s", string(data))
	}
}

func TestConcurrentHealth(t *testing.T) {
	ct := time.Date(2020, 12, 16, 15, 30, 46, 0, time.UTC)

	tests := []struct {
		name      string
		checkTime time.Time
		content   string
		want      bool
	}{
		{
			name:      "after",
			checkTime: ct,
			content:   "2020-12-16T15:30:45Z",
			want:      true,
		},
		{
			name:      "before",
			checkTime: ct,
			content:   "2020-12-16T15:30:47Z",
			want:      true,
		},
		{
			name:      "far before",
			checkTime: ct,
			content:   "2020-12-16T15:29:45Z",
			want:      false,
		},
		{
			name:      "far after",
			checkTime: ct,
			content:   "2020-12-16T15:31:45Z",
			want:      false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, wasConcurrentHealthCheck(tc.checkTime, tc.content))
		})
	}
}
