package objstore

import (
	"bytes"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"golang.org/x/xerrors"
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
func (h *HealthCheck) HealthCheck() servicehealthcheck.HealthCheckResult {
	if time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		// the last health check is not outdated, an can be reused.
		return h.state.GetState()
	}

	expContent := []byte(time.Now().Format(time.RFC3339))
	expSize := int64(len(expContent))

	// Try upload
	_, err := h.Client.PutObject(
		cfg.HealthCheckBucketName,
		cfg.HealthCheckObjectName,
		bytes.NewReader(expContent),
		expSize,
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	// Try download
	obj, err := h.Client.GetObject(
		cfg.HealthCheckBucketName,
		cfg.HealthCheckObjectName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	// Assert expectations
	gotContent := make([]byte, expSize)
	_, err = obj.Read(gotContent)
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	defer obj.Close()

	if bytes.Compare(gotContent, expContent) == 0 {
		h.state.SetErrorState(xerrors.New("objstore: unexpected health check caused by unexpected object content"))
		return h.state.GetState()
	}

	// If uploading and downloading worked set the Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()
}
