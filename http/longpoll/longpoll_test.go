package longpoll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLongPollUntilBounds(t *testing.T) {
	called := 0
	ok, err := Until(context.Background(), -1, func(ctx context.Context) (bool, error) {
		budget, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.Equal(t, time.Millisecond*999, budget.Sub(time.Now()).Truncate(time.Millisecond)) // nolint: gosimple
		called++
		return true, nil
	})
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)

	called = 0
	ok, err = Until(context.Background(), time.Hour, func(ctx context.Context) (bool, error) {
		budget, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.Equal(t, time.Second*59, budget.Sub(time.Now()).Truncate(time.Second)) // nolint: gosimple
		called++
		return true, nil
	})
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestLongPollUntilNoTimeout(t *testing.T) {
	called := 0
	ok, err := Until(context.Background(), 0, func(context.Context) (bool, error) {
		called++
		return false, nil
	})
	assert.False(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)

	called = 0
	ok, err = Until(context.Background(), 0, func(context.Context) (bool, error) {
		called++
		return true, nil
	})
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestLongPollUntilErr(t *testing.T) {
	called := 0
	ok, err := Until(context.Background(), 0, func(context.Context) (bool, error) {
		called++
		return false, errors.New("Foo")
	})
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Equal(t, 1, called)
}

func TestLongPollUntilTimeout(t *testing.T) {
	called := 0
	ok, err := Until(context.Background(), time.Second, func(context.Context) (bool, error) {
		called++
		return false, nil
	})
	assert.False(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestLongPollUntilTimeoutWithContext(t *testing.T) {
	called := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ok, err := Until(ctx, time.Second*2, func(context.Context) (bool, error) {
		called++
		return false, nil
	})
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.GreaterOrEqual(t, called, 2)
	assert.LessOrEqual(t, called, 3)
}
