package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIntegration_RedisLock(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()

	lock := NewLock("test")

	lockCtx, releaseLock, err := lock.AcquireAndKeepUp(ctx)
	require.NoError(t, err)
	require.NotNil(t, lockCtx)
	require.NotNil(t, releaseLock)
	releaseLock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for try := 0; true; try++ {
		lockCtx, releaseLock, err = lock.AcquireAndKeepUp(ctx)
		require.NoError(t, err)

		if lockCtx == nil {
			t.Log("Not obtained, try again in 1sec")
			time.Sleep(time.Second)

			continue
		}

		require.NotNil(t, lockCtx)
		require.NotNil(t, releaseLock)
		releaseLock()

		break
	}
}
