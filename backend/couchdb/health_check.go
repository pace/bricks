package couchdb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	kivik "github.com/go-kivik/kivik/v3"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

// HealthCheck checks the state of the object storage client. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	Name   string
	Client *kivik.Client
	DB     *kivik.DB
	Config *Config

	state servicehealthcheck.ConnectionState
}

var (
	healthCheckTimeFormat     = time.RFC3339
	healthCheckConcurrentSpan = 10 * time.Second
)

// HealthCheck checks if the object storage client is healthy. If the last result is outdated,
// object storage is checked for upload and download,
// otherwise returns the old result
func (h *HealthCheck) HealthCheck(ctx context.Context) servicehealthcheck.HealthCheckResult {
	if time.Since(h.state.LastChecked()) <= h.Config.HealthCheckResultTTL {
		// the last health check is not outdated, an can be reused.
		return h.state.GetState()
	}

	checkTime := time.Now()

	var doc Doc
	var err error
	var row *kivik.Row

check:
	// check if context was canceled
	select {
	case <-ctx.Done():
		h.state.SetErrorState(fmt.Errorf("failed: %v", ctx.Err()))
		return h.state.GetState()
	default:
	}

	row = h.DB.Get(ctx, h.Config.HealthCheckKey)
	if row.Err != nil {
		if kivik.StatusCode(row.Err) == http.StatusNotFound {
			goto put
		}
		h.state.SetErrorState(fmt.Errorf("failed to get: %#v", row.Err))
		return h.state.GetState()
	}
	defer row.Body.Close()

	// check if document exists
	if row.Rev != "" {
		err = row.ScanDoc(&doc)
		if err != nil {
			h.state.SetErrorState(fmt.Errorf("failed to get: %v", row.Err))
			return h.state.GetState()
		}

		// check was already perfromed
		if wasConcurrentHealthCheck(checkTime, doc.Time) {
			goto healthy
		}
	}

put:
	// update document
	doc.ID = h.Config.HealthCheckKey
	doc.Time = time.Now().Format(healthCheckTimeFormat)
	_, err = h.DB.Put(ctx, h.Config.HealthCheckKey, doc)
	if err != nil {
		// not yet created, try to create
		if h.Config.DatabaseAutoCreate && kivik.StatusCode(err) == http.StatusNotFound {
			err := h.Client.CreateDB(ctx, h.Name)
			if err != nil {
				h.state.SetErrorState(fmt.Errorf("failed to put object: %v", err))
				return h.state.GetState()
			}
			goto put
		}

		if kivik.StatusCode(err) == http.StatusConflict {
			goto check
		}
		h.state.SetErrorState(fmt.Errorf("failed to put object: %v", err))
		return h.state.GetState()
	}

	// document was uploaded goto check
	goto check

healthy:
	// If uploading and downloading worked set the Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()
}

type Doc struct {
	ID   string `json:"_id"`
	Rev  string `json:"_rev,omitempty"`
	Time string `json:"at"`
}

// wasConcurrentHealthCheck checks if the time doesn't match in a certain
// time span concurrent request to the objstore may break the assumption
// that the value is the same, but in this case it would be acceptable.
// Assumption all instances are created equal and one providing evidence
// of a good write would be sufficient. See #244
func wasConcurrentHealthCheck(checkTime time.Time, observedValue string) bool {
	t, err := time.Parse(healthCheckTimeFormat, observedValue)
	if err == nil {
		allowedStart := checkTime.Add(-healthCheckConcurrentSpan)
		allowedEnd := checkTime.Add(healthCheckConcurrentSpan)

		// timestamp we got from the document is in allowed range
		// concider it healthy
		return t.After(allowedStart) && t.Before(allowedEnd)
	}

	return false
}
