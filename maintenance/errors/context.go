package errors

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Hide checks whether the given error err is a context.Canceled or
// context.DeadlineExceeded error and if so exposes these errors as
// wrapped errors while maintaining the given err as string. If
// the optional exposedErr is provided, it is attached as prefix
// to the returned error. If the given err is not any of the above
// ones, the given exposedErr (if present) is wrapped as prefix
// to the returned error.
func Hide(ctx context.Context, err, exposedErr error) error {
	ret := errors.New(err.Error())
	if exposedErr != nil {
		ret = fmt.Errorf("%w: %s", exposedErr, err)
	}

	if ctx.Err() == context.Canceled && errors.Is(err, context.Canceled) {
		s := strings.TrimSuffix(err.Error(), context.Canceled.Error())
		return fmt.Errorf("%s%w", s, context.Canceled)
	}

	if ctx.Err() == context.DeadlineExceeded && errors.Is(err, context.DeadlineExceeded) {
		s := strings.TrimSuffix(err.Error(), context.DeadlineExceeded.Error())
		return fmt.Errorf("%s%w", s, context.DeadlineExceeded)
	}

	return ret
}
