// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package synctx

import (
	"sync"
)

// WaitGroup extended with Finish func.
type WaitGroup struct {
	sync.WaitGroup
}

// Finish allows to be used easily with go contexts.
func (wg *WaitGroup) Finish() <-chan struct{} {
	ch := make(chan struct{})
	go func() { wg.Wait(); close(ch) }()

	return ch
}
