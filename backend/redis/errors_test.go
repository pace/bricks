package redis_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	pbredis "github.com/pace/bricks/backend/redis"
)

func TestIsErrConnectionFailed(t *testing.T) {
	t.Run("connection failed (io.EOF)", func(t *testing.T) {
		err := fmt.Errorf("%w", io.EOF)
		require.True(t, pbredis.IsErrConnectionFailed(err))
	})

	t.Run("connection failed (net.Error)", func(t *testing.T) {
		c := pbredis.CustomClient(&redis.Options{}) // invalid connection
		err := c.Ping(context.Background()).Err()
		require.True(t, pbredis.IsErrConnectionFailed(err))
	})

	t.Run("any other error", func(t *testing.T) {
		err := errors.New("any other error")
		require.False(t, pbredis.IsErrConnectionFailed(err))
	})
}
