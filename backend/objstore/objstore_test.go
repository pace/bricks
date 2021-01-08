package objstore

import (
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	client, err := Client()

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestCustomClient(t *testing.T) {
	client, err := CustomClient("s3.amazonaws.com", &minio.Options{
		Region: "eu-central-1",
	})

	require.NoError(t, err)
	assert.NotNil(t, client)
}
