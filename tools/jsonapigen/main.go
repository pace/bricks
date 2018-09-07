// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/30 by Vincent Landgraf

package main

import (
	"flag"
	"os"
	"path/filepath"

	"lab.jamit.de/pace/go-microservice/http/jsonapi/generator"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

var pkg, path, source string

func main() {
	flag.StringVar(&pkg, "pkg", pkg, "go package name")
	flag.StringVar(&path, "path", path, "path for generated file")
	flag.StringVar(&source, "source", source, "source OpenAPIv3 document")
	flag.Parse()

	var g generator.Generator

	s, err := g.BuildSource(source, filepath.Dir(pkg), filepath.Base(pkg))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString(s)
	if err != nil {
		log.Fatal(err)
	}
}
