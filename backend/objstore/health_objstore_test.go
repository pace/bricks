package objstore

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	http2 "github.com/pace/bricks/http"
	"github.com/pace/bricks/maintenance/log"
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

// TestIntegrationHealthCheck tests if object storage health check ist working like expected
func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	RegisterHealthchecks()
	resp := setup()
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health/check to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(string(data), "objstore               OK") {
		t.Errorf("Expected /health/check to return OK, got: %q", string(data[:]))
	}
}
