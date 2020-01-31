package objstore

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

// HealthCheck checks the state of the object storage client. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	state  servicehealthcheck.ConnectionState
	Client *minio.Client
}

// HealthCheck checks if the object storage client is healthy. If the last result is outdated,
// object storage is checked for upload and download,
// otherwise returns the old result
func (h *HealthCheck) HealthCheck(ctx context.Context) servicehealthcheck.HealthCheckResult {
	if time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		// the last health check is not outdated, an can be reused.
		return h.state.GetState()
	}

	expContent := []byte(time.Now().Format(time.RFC3339))
	expSize := int64(len(expContent))

	_, err := h.Client.PutObjectWithContext(
		ctx,
		cfg.HealthCheckBucketName,
		cfg.HealthCheckObjectName,
		bytes.NewReader(expContent),
		expSize,
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	if err != nil {
		h.state.SetErrorState(fmt.Errorf("failed to put object: %v", err))
		return h.state.GetState()
	}

	// Try download
	obj, err := h.Client.GetObjectWithContext(
		ctx,
		cfg.HealthCheckBucketName,
		cfg.HealthCheckObjectName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		h.state.SetErrorState(fmt.Errorf("failed to get object: %v", err))
		return h.state.GetState()
	}
	defer obj.Close()

	// Assert expectations
	buf, err := ioutil.ReadAll(obj)
	if err != nil {
		h.state.SetErrorState(fmt.Errorf("failed to compare object: %v", err))
		return h.state.GetState()
	}

	if !bytes.Equal(buf, expContent) {
		h.state.SetErrorState(fmt.Errorf("unexpected content: %q <-> %q", string(buf), string(expContent)))
		return h.state.GetState()
	}

	// If uploading and downloading worked set the Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()
}
