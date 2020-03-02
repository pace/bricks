// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/27 by Marius Neugebauer

package routine

import (
	"time"

	"github.com/caarlos0/env"
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
