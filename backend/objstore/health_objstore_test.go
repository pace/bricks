package objstore

import (
	http2 "github.com/pace/bricks/http"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

import (
	"github.com/pace/bricks/maintenance/log"
	"strings"
)

func setup() *http.Response {
	r := http2.Router()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/check", nil)
	r.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	return resp
}

// TestIntegrationHealthCheck tests if redis health check ist working like expected
func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	resp := setup()
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health/check to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(string(data), "redis                  OK") {
		t.Errorf("Expected /health/check to return OK, got: %q", string(data[:]))
	}
}
