// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package cache_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pace/bricks/pkg/cache"
)

func Example_inMemory() {
	ctx := context.Background()

	// init cache
	var c cache.Cache = cache.InMemory()

	// write to cache
	if err := c.Put(ctx, "foo", []byte("bar"), time.Hour); err != nil {
		panic(err)
	}

	// get from cache and print
	v, _, err := c.Get(ctx, "foo")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(v))

	// forget
	if err := c.Forget(ctx, "foo"); err != nil {
		panic(err)
	}

	// get from cache and print
	_, _, err = c.Get(ctx, "foo")
	if errors.Is(err, cache.ErrNotFound) {
		fmt.Println(err)
	} else {
		panic("expected error not found")
	}

	// Output:
	// bar
	// key "foo": not found
}
