// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/08/12 by Marius Neugebauer

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
