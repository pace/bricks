// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package cache_test

import (
	"testing"

	"github.com/pace/bricks/pkg/cache"
	"github.com/pace/bricks/pkg/cache/testsuite"
	"github.com/stretchr/testify/suite"
)

func TestMemory(t *testing.T) {
	suite.Run(t, &testsuite.CacheTestSuite{
		Cache: cache.InMemory(),
	})
}
