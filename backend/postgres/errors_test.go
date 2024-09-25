package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	pbpostgres "github.com/pace/bricks/backend/postgres"
)

func TestIsErrConnectionFailed(t *testing.T) {
	t.Run("connection failed (io.EOF)", func(t *testing.T) {
		err := fmt.Errorf("%w", io.EOF)
		require.True(t, pbpostgres.IsErrConnectionFailed(err))
	})

	t.Run("connection failed (net.Error)", func(t *testing.T) {
		ctx := context.Background()

		db := pbpostgres.NewDB(ctx, pbpostgres.WithHost("foobar")) // invalid connection
		_, err := db.Exec("")
		require.True(t, pbpostgres.IsErrConnectionFailed(err))
	})

	t.Run("any other error", func(t *testing.T) {
		err := errors.New("any other error")
		require.False(t, pbpostgres.IsErrConnectionFailed(err))
	})
}
