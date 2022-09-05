package errors

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHide(t *testing.T) {
	iAmAnError := errors.New("i am an error")
	iAmAnotherError := errors.New("i am another error")

	backgroundContext := context.Background()
	canceledContext, cancel := context.WithCancel(backgroundContext)
	cancel()

	exceededContext, cancel2 := context.WithTimeout(backgroundContext, time.Millisecond)
	defer cancel2()
	time.Sleep(2 * time.Millisecond)

	type args struct {
		ctx        context.Context
		err        error
		exposedErr error
	}
	tests := []struct {
		name             string
		args             args
		want             error
		expectContextErr bool
	}{
		{
			name: "normal_context_no_error_nothing_exposed",
			args: args{
				ctx:        backgroundContext,
				err:        iAmAnError,
				exposedErr: nil,
			},
			want: iAmAnError,
		},
		{
			name: "normal_context_no_error_with_exposed",
			args: args{
				ctx:        backgroundContext,
				err:        iAmAnError,
				exposedErr: iAmAnotherError,
			},
			want: fmt.Errorf("%w: %s", iAmAnotherError, iAmAnError),
		},
		{
			name: "normal_context_with_error_nothing_exposed",
			args: args{
				ctx:        backgroundContext,
				err:        fmt.Errorf("%s: %w", iAmAnError, context.Canceled),
				exposedErr: nil,
			},
			want:             fmt.Errorf("%s: %s", iAmAnError, context.Canceled),
			expectContextErr: true,
		},
		{
			name: "normal_context_with_error_with_exposed",
			args: args{
				ctx:        backgroundContext,
				err:        fmt.Errorf("%s: %w", iAmAnError, context.Canceled),
				exposedErr: iAmAnotherError,
			},
			want:             fmt.Errorf("%w: %s: %s", iAmAnotherError, iAmAnError, context.Canceled),
			expectContextErr: true,
		},
		{
			name: "canceled_context_no_error_nothing_exposed",
			args: args{
				ctx:        canceledContext,
				err:        iAmAnError,
				exposedErr: nil,
			},
			want:             iAmAnError,
			expectContextErr: false,
		},
		{
			name: "canceled_context_no_error_with_exposed",
			args: args{
				ctx:        canceledContext,
				err:        iAmAnError,
				exposedErr: iAmAnotherError,
			},
			want:             fmt.Errorf("%w: %s", iAmAnotherError, iAmAnError),
			expectContextErr: false,
		},
		{
			name: "canceled_context_with_error_nothing_exposed",
			args: args{
				ctx:        canceledContext,
				err:        fmt.Errorf("%s: %w", iAmAnError, context.Canceled),
				exposedErr: nil,
			},
			want:             fmt.Errorf("%s: %w", iAmAnError, context.Canceled),
			expectContextErr: true,
		},
		{
			name: "canceled_context_with_error_with_exposed",
			args: args{
				ctx:        canceledContext,
				err:        fmt.Errorf("%s: %w", iAmAnError, context.Canceled),
				exposedErr: iAmAnotherError,
			},
			want:             fmt.Errorf("%s: %s: %w", iAmAnotherError, iAmAnError, context.Canceled),
			expectContextErr: true,
		},
		{
			name: "exceeded_context_no_error_nothing_exposed",
			args: args{
				ctx:        exceededContext,
				err:        iAmAnError,
				exposedErr: nil,
			},
			want:             iAmAnError,
			expectContextErr: false,
		},
		{
			name: "exceeded_context_no_error_with_exposed",
			args: args{
				ctx:        exceededContext,
				err:        iAmAnError,
				exposedErr: iAmAnotherError,
			},
			want:             fmt.Errorf("%w: %s", iAmAnotherError, iAmAnError),
			expectContextErr: false,
		},
		{
			name: "exceeded_context_with_error_nothing_exposed",
			args: args{
				ctx:        exceededContext,
				err:        fmt.Errorf("%s: %w", iAmAnError, context.DeadlineExceeded),
				exposedErr: nil,
			},
			want:             fmt.Errorf("%s: %w", iAmAnError, context.DeadlineExceeded),
			expectContextErr: true,
		},
		{
			name: "exceeded_context_with_error_with_exposed",
			args: args{
				ctx:        exceededContext,
				err:        fmt.Errorf("%s: %w", iAmAnError, context.DeadlineExceeded),
				exposedErr: iAmAnotherError,
			},
			want:             fmt.Errorf("%s: %s: %w", iAmAnotherError, iAmAnError, context.DeadlineExceeded),
			expectContextErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Hide(tt.args.ctx, tt.args.err, tt.args.exposedErr)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.expectContextErr, IsStdLibContextError(tt.args.err))
		})
	}
}
