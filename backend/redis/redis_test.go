// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package redis

import (
	"context"
	"testing"
)

func TestRedisClient(t *testing.T) {
	c := WithContext(context.Background(), Client())
	c.Ping()
}

func TestRedisClusterClient(t *testing.T) {
	c := WithClusterContext(context.Background(), ClusterClient())
	c.Ping()
}
