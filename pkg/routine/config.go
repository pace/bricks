// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package routine

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type config struct {
	RedisLockTTL time.Duration `env:"ROUTINE_REDIS_LOCK_TTL" envDefault:"5s"`
}

var cfg config

func init() {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
}
