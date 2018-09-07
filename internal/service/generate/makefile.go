// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/31 by Vincent Landgraf

package generate

import (
	"html/template"
	"log"
	"os"
)

// MakefileOptions options that change the rendering
// of the makefile
type MakefileOptions struct {
	Name string
}

// Makefile generates a with given options for the
// specified path
func Makefile(path string, options MakefileOptions) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	err = makefileTemplate.Execute(f, options)
	if err != nil {
		log.Fatal(err)
	}
}

var makefileTemplate = template.Must(template.New("Makefile").Parse(`.PHONY: docker

docker:
	docker build -t {{ .Name }} .
`))
