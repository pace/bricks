// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package cache_test

import (
	"testing"

	"github.com/pace/bricks/backend/redis"
	"github.com/pace/bricks/pkg/cache"
	"github.com/pace/bricks/pkg/cache/testsuite"
	"github.com/stretchr/testify/suite"
)

func TestIntegrationRedis(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	suite.Run(t, &testsuite.CacheTestSuite{
		Cache: cache.InRedis(redis.Client(), "test:cache:"),
	})
}
