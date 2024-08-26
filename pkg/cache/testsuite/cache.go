// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package testsuite

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/cache"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	Cache cache.Cache
}

func (suite *CacheTestSuite) TestPut() {
	c := suite.Cache
	ctx := log.WithContext(context.Background())

	_ = c.Forget(ctx, "foo") // make sure it doesn't exist
	suite.Run("does not error", func() {
		err := c.Put(ctx, "foo", []byte("bar"), time.Second)
		suite.NoError(err)
	})
	_ = c.Forget(ctx, "foo") // clean up

	_ = c.Forget(ctx, "") // make sure it doesn't exist
	suite.Run("accepts all null values", func() {
		err := c.Put(ctx, "", nil, 0)
		suite.NoError(err)
	})
	_ = c.Forget(ctx, "") // clean up

	_ = c.Forget(ctx, "中文پنجابی🥰🥸") // make sure it doesn't exist
	suite.Run("supports unicode", func() {
		err := c.Put(ctx, "中文پنجابی🥰🥸", []byte("🦤ᐃᓄᒃᑎᑐᑦລາວ"), 0)
		suite.NoError(err)
	})
	_ = c.Forget(ctx, "中文پنجابی🥰🥸") // clean up

	suite.Run("does not error when repeated", func() {
		_ = c.Put(ctx, "foo", []byte("bar"), time.Second)
		err := c.Put(ctx, "foo", []byte("bar"), time.Second)
		suite.NoError(err)
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("stores a value", func() {
		_ = c.Put(ctx, "foo", []byte("bar"), 0)
		value, _, _ := c.Get(ctx, "foo")
		suite.Equal([]byte("bar"), value)
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("is unaffected from manipulating the input", func() {
		input := []byte("bar")
		_ = c.Put(ctx, "foo", input, 0)
		input[0]++ // input manipulation
		value, _, _ := c.Get(ctx, "foo")
		suite.Equal([]byte("bar"), value)
	})
	_ = c.Forget(ctx, "foo") // clean up

	for i := 0; i <= 5; i++ { // make sure it doesn't exist
		_ = c.Forget(ctx, fmt.Sprintf("foo%d", i))
	}
	suite.Run("does not error on simultaneous use", func() {
		var wg sync.WaitGroup
		for i := 0; i <= 5; i++ {
			wg.Add(1)
			go func() {
				err := c.Put(ctx, fmt.Sprintf("foo%d", i), []byte("bar"), 0)
				suite.NoError(err)
				wg.Done()
			}()
			wg.Wait()
		}
	})
	for i := 0; i <= 5; i++ { // clean up
		_ = c.Forget(ctx, fmt.Sprintf("foo%d", i))
	}
}

func (suite *CacheTestSuite) TestGet() {
	c := suite.Cache
	ctx := log.WithContext(context.Background())

	_ = c.Forget(ctx, "foo") // make sure it doesn't exist
	suite.Run("returns the ttl if set", func() {
		_ = c.Put(ctx, "foo", []byte("bar"), time.Minute)
		_, ttl, _ := c.Get(ctx, "foo")
		// expect ttl to be something between 59 and 60 seconds
		suite.LessOrEqual(int64(ttl), int64(time.Minute))
		suite.Greater(int64(ttl), int64(time.Minute-time.Second))
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("returns 0 as ttl if ttl not set", func() {
		_ = c.Put(ctx, "foo", []byte("bar"), 0)
		_, ttl, _ := c.Get(ctx, "foo")
		suite.Equal(time.Duration(0), ttl)
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("returns not found error", func() {
		_, _, err := c.Get(ctx, "foo")
		suite.True(errors.Is(err, cache.ErrNotFound))
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("returns not found if ttl ran out", func() {
		err := c.Put(ctx, "foo", []byte("bar"), time.Millisecond) // minimum ttl
		suite.NoError(err)
		<-time.After(2 * time.Millisecond)
		_, _, err = c.Get(ctx, "foo")
		suite.True(errors.Is(err, cache.ErrNotFound))
	})
	_ = c.Forget(ctx, "foo") // clean up

	_ = c.Forget(ctx, "foo1") // make sure it doesn't exist
	_ = c.Forget(ctx, "foo2") // make sure it doesn't exist
	suite.Run("retrieves the right value", func() {
		_ = c.Put(ctx, "foo1", []byte("bar1"), 0)
		_ = c.Put(ctx, "foo2", []byte("bar2"), 0)
		value1, _, _ := c.Get(ctx, "foo1")
		value2, _, _ := c.Get(ctx, "foo2")
		suite.Equal([]byte("bar1"), value1)
		suite.Equal([]byte("bar2"), value2)
	})
	_ = c.Forget(ctx, "foo1") // clean up
	_ = c.Forget(ctx, "foo2") // clean up

	suite.Run("is unaffected from manipulating the output", func() {
		_ = c.Put(ctx, "foo", []byte("bar"), 0)
		output, _, _ := c.Get(ctx, "foo")
		output[0]++ // output manipulation
		value, _, _ := c.Get(ctx, "foo")
		suite.Equal([]byte("bar"), value)
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("does not produce nil", func() {
		_ = c.Put(ctx, "foo", nil, 0)
		value, _, _ := c.Get(ctx, "foo")
		suite.NotNil(value)
	})
	_ = c.Forget(ctx, "foo") // clean up

	_ = c.Forget(ctx, "") // make sure it doesn't exist
	suite.Run("returns value stored with an empty key", func() {
		_ = c.Put(ctx, "", []byte("bar"), 0)
		value, _, _ := c.Get(ctx, "")
		suite.Equal([]byte("bar"), value)
	})
	_ = c.Forget(ctx, "") // clean up

	_ = c.Forget(ctx, "中文پنجابی🥰🥸") // make sure it doesn't exist
	suite.Run("supports unicode", func() {
		_ = c.Put(ctx, "中文پنجابی🥰🥸", []byte("🦤ᐃᓄᒃᑎᑐᑦລາວ\x00"), 0)
		value, _, _ := c.Get(ctx, "中文پنجابی🥰🥸")
		suite.Equal([]byte("🦤ᐃᓄᒃᑎᑐᑦລາວ\x00"), value)
	})
	_ = c.Forget(ctx, "中文پنجابی🥰🥸") // clean up

	for i := 0; i <= 5; i++ { // make sure it doesn't exist
		_ = c.Forget(ctx, fmt.Sprintf("foo%d", i))
	}
	suite.Run("does not error on simultaneous use", func() {
		for i := 0; i <= 5; i++ {
			_ = c.Put(ctx, fmt.Sprintf("foo%d", i), []byte("bar"), 0)
		}
		var wg sync.WaitGroup
		for i := 0; i <= 5; i++ {
			wg.Add(1)
			go func() {
				_, _, err := c.Get(ctx, fmt.Sprintf("foo%d", i))
				suite.NoError(err)
				wg.Done()
			}()
			wg.Wait()
		}
	})
	for i := 0; i <= 5; i++ { // clean up
		_ = c.Forget(ctx, fmt.Sprintf("foo%d", i))
	}
}

func (suite *CacheTestSuite) TestForget() {
	c := suite.Cache
	ctx := log.WithContext(context.Background())

	_ = c.Forget(ctx, "foo") // make sure it doesn't exist
	suite.Run("works", func() {
		_ = c.Put(ctx, "foo", []byte("bar"), 0)
		_ = c.Forget(ctx, "foo")
		_, _, err := c.Get(ctx, "foo")
		suite.True(errors.Is(err, cache.ErrNotFound))
	})
	_ = c.Forget(ctx, "foo") // clean up

	suite.Run("does not error when repeated", func() {
		_ = c.Forget(ctx, "foo")
		err := c.Forget(ctx, "foo")
		suite.NoError(err)
	})
	_ = c.Forget(ctx, "foo") // clean up

	_ = c.Forget(ctx, "中文پنجابی🥰🥸") // make sure it doesn't exist
	suite.Run("supports unicode", func() {
		err := c.Forget(ctx, "中文پنجابی🥰🥸")
		suite.NoError(err)
	})
	_ = c.Forget(ctx, "中文پنجابی🥰🥸") // clean up

	for i := 0; i <= 5; i++ { // make sure it doesn't exist
		_ = c.Forget(ctx, fmt.Sprintf("foo%d", i))
	}
	suite.Run("does not error on simultaneous use", func() {
		for i := 0; i <= 5; i++ {
			_ = c.Put(ctx, fmt.Sprintf("foo%d", i), []byte("bar"), 0)
		}
		var wg sync.WaitGroup
		for i := 0; i <= 5; i++ {
			wg.Add(1)
			go func() {
				err := c.Forget(ctx, fmt.Sprintf("foo%d", i))
				suite.NoError(err)
				wg.Done()
			}()
			wg.Wait()
		}
	})
	for i := 0; i <= 5; i++ { // clean up
		_ = c.Forget(ctx, fmt.Sprintf("foo%d", i))
	}
}
