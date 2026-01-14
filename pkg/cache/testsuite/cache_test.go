// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

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
