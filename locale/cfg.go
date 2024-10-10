// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package locale

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	Language string `env:"LOCALE_FALLBACK_LANGUAGE" envDefault:"de-DE"`
	Timezone string `env:"LOCALE_FALLBACK_TIMEZONE" envDefault:"Europe/Berlin"`
}

var cfg config

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse environment: %v", err)
	}
}
