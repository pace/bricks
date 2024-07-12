// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package testsuite_test

import (
	"testing"

	"github.com/pace/bricks/pkg/cache"
	. "github.com/pace/bricks/pkg/cache/testsuite"
	"github.com/stretchr/testify/suite"
)

// TestStringsTestSuite tests the reference in-memory cache implementation.
func TestStringsTestSuite(t *testing.T) {
	suite.Run(t, &CacheTestSuite{
		Cache: cache.InMemory(),
	})
}
