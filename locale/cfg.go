// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/04/16 by Vincent Landgraf

package locale

import (
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	Language string `env:"LOCALE_FALLBACK_LANGUAGE" envDefault:"de-DE"`
	Timezone string `env:"LOCALE_FALLBACK_TIMEZONE" envDefault:"Europe/Berlin"`
}

var cfg config

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse environment: %v", err)
	}
}
