package transport

import "errors"

var (
	ErrRetryFailed   = errors.New("failed after maximum number of retries")
	ErrCircuitBroken = errors.New("circuit to remote host is open")
)
