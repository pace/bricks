package objstore

import (
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient(t *testing.T) {
	client, err := Client()

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestCustomClient(t *testing.T) {
	client, err := CustomClient("example.org", &minio.Options{
		Region: "eu-central-1",
	})

	require.NoError(t, err)
	assert.NotNil(t, client)
}
