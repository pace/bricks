// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package generate

import (
	"log"
	"os"

	"github.com/pace/bricks/http/jsonapi/generator"
)

// RestOptions options to respect when generating the rest api
type RestOptions struct {
	PkgName, Path, Source string
}

// Rest builds a jsonapi rest api
func Rest(options RestOptions) {
	// generate jsonapi
	g := generator.Generator{}
	result, err := g.BuildSource(options.Source, options.Path, options.PkgName)
	if err != nil {
		log.Fatal(err)
	}

	// create file
	file, err := os.Create(options.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close() // nolint: errcheck

	// write file
	_, err = file.WriteString(result)
	if err != nil {
		log.Fatal(err)
	}
}
