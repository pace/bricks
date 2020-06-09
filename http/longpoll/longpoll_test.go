package longpoll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLongPollUntilNoTimeout(t *testing.T) {
	called := 0
	ok, err := LongPollUntil(context.Background(), 0, func() (bool, error) {
		called++
		return false, nil
	})
	assert.False(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)

	called = 0
	ok, err = LongPollUntil(context.Background(), 0, func() (bool, error) {
		called++
		return true, nil
	})
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestLongPollUntilErr(t *testing.T) {
	called := 0
	ok, err := LongPollUntil(context.Background(), 0, func() (bool, error) {
		called++
		return false, errors.New("Foo")
	})
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Equal(t, 1, called)
}

func TestLongPollUntilTimeout(t *testing.T) {
	called := 0
	ok, err := LongPollUntil(context.Background(), time.Second*1, func() (bool, error) {
		called++
		return false, nil
	})
	assert.False(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, 3, called)
}

func TestLongPollUntilTimeoutWithContext(t *testing.T) {
	called := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ok, err := LongPollUntil(ctx, time.Second*2, func() (bool, error) {
		called++
		return false, nil
	})
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.GreaterOrEqual(t, called, 2)
	assert.LessOrEqual(t, called, 3)
}
