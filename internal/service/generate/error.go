// Copyright © 2023 by PACE Telematics GmbH. All rights reserved.

package generate

import (
	"log"
	"os"

	"github.com/pace/bricks/internal/service/generate/errordefinition/generator"
)

// ErrorDefinitionFileOptions options that change the rendering of the error definition file
type ErrorDefinitionFileOptions struct {
	PkgName, Path, Source string
}

// ErrorDefinitionFile builds a file with error definitions
func ErrorDefinitionFile(options ErrorDefinitionFileOptions) {

	// generate error definition
	g := generator.Generator{}
	result, err := g.BuildSource(options.Source, options.Path, options.PkgName)
	if err != nil {
		log.Fatal(err)
	}

	writeResult(result, options.Path)
}

func ErrorDefinitionsMarkdown(options ErrorDefinitionFileOptions) {
	g := generator.Generator{}
	result, err := g.BuildMarkdown(options.Source)
	if err != nil {
		log.Fatal(err)
	}

	writeResult(result, options.Path)
}

func writeResult(result, path string) {
	// create file
	file, err := os.Create(path)
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
