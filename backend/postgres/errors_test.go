package postgres_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/require"

	pbpostgres "github.com/pace/bricks/backend/postgres"
)

func TestIsErrConnectionFailed(t *testing.T) {
	t.Run("connection failed (io.EOF)", func(t *testing.T) {
		err := fmt.Errorf("%w", io.EOF)
		require.True(t, pbpostgres.IsErrConnectionFailed(err))
	})

	t.Run("connection failed (net.Error)", func(t *testing.T) {
		db := pbpostgres.CustomConnectionPool(&pg.Options{}) // invalid connection
		_, err := db.Exec("")
		require.True(t, pbpostgres.IsErrConnectionFailed(err))
	})

	t.Run("connection failed (pg.Error)", func(t *testing.T) {
		err := error(mockPGError{m: map[byte]string{'C': "08000"}})
		require.True(t, pbpostgres.IsErrConnectionFailed(err))
	})

	t.Run("any other error", func(t *testing.T) {
		err := errors.New("any other error")
		require.False(t, pbpostgres.IsErrConnectionFailed(err))
	})
}

type mockPGError struct {
	m map[byte]string
}

func (err mockPGError) Field(k byte) string      { return err.m[k] }
func (err mockPGError) IntegrityViolation() bool { return false }
func (err mockPGError) Error() string            { return fmt.Sprintf("%+v", err.m) }
