// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package objstore

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestObjStoreClient(t *testing.T) {
	c, err := Client()
	require.NoError(t, err)
	assert.NotNil(t, c)
}
