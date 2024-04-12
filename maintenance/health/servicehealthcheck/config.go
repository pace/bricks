package servicehealthcheck

import (
	"time"

	"github.com/caarlos0/env/v10"

	"github.com/pace/bricks/maintenance/log"
)

// config is the global/default config.
type config struct {
	// Amount of time to wait until next health check
	Interval time.Duration `env:"HEALTH_CHECK_INTERVAL" envDefault:"1m"`
	// Amount of time to cache the last init
	HealthCheckInitResultErrorTTL time.Duration `env:"HEALTH_CHECK_INIT_RESULT_ERROR_TTL" envDefault:"10s"`
	// Amount of time to wait before failing the health check
	HealthCheckMaxWait time.Duration `env:"HEALTH_CHECK_MAX_WAIT" envDefault:"5s"`
	// Amount of time given to warmup before running the first full healtheck. Delay will be applied based on the
	// healthcheck start time and not delay a possible required initialization.
	HealthCheckWarmupDelay time.Duration `env:"HEALTH_CHECK_WARMUP_DELAY" envDefault:"0s"`
}

var cfg config

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse health check environment: %v", err)
	}
}

// HealthCheckCfg is the config used per HealthCheck.
type HealthCheckCfg struct {
	interval           time.Duration
	initResultErrorTTL time.Duration
	maxWait            time.Duration
	warmupDelay        time.Duration
}

type HealthCheckOption func(cfg *HealthCheckCfg)

func UseInterval(interval time.Duration) HealthCheckOption {
	return func(cfg *HealthCheckCfg) {
		cfg.interval = interval
	}
}

func UseInitErrResultTTL(ttl time.Duration) HealthCheckOption {
	return func(cfg *HealthCheckCfg) {
		cfg.initResultErrorTTL = ttl
	}
}

func UseMaxWait(maxWait time.Duration) HealthCheckOption {
	return func(cfg *HealthCheckCfg) {
		cfg.maxWait = maxWait
	}
}

// UseWarmup - delays a healthcheck during warmup
func UseWarmup(delay time.Duration) HealthCheckOption {
	return func(cfg *HealthCheckCfg) {
		cfg.warmupDelay = delay
	}
}
