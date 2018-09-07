// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/31 by Vincent Landgraf

package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"lab.jamit.de/pace/go-microservice/internal/service/generate"
)

// NewOptions collection of options to apply while or
// after the creation of the new project
type NewOptions struct {
	RestSource string // url or path to OpenAPIv3 (json:api) specification
}

// New creates a new directory in the go path
func New(name string, options NewOptions) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	SimpleExec("git", "init", dir)
	SimpleExecInPath(dir, "git", "remote", "add", "origin", fmt.Sprintf(GitLabTemplate, name))
	log.Printf("Remember to create the %s repository in gitlab: https://lab.jamit.de/projects/new?namespace_id=296\n", name)

	// add REST API if there was a source specified
	if options.RestSource != "" {
		restDir := filepath.Join(dir, "internal", "http", "rest")
		err := os.MkdirAll(restDir, 0770)
		if err != nil {
			log.Fatal(fmt.Printf("Failed to generate dir for rest api %s: %v", restDir, err))
		}

		generate.Rest(generate.RestOptions{
			Path:    filepath.Join(restDir, "jsonapi.go"),
			PkgName: "rest",
			Source:  options.RestSource,
		})
	}

	SimpleExecInPath(dir, "go", "mod", "init", GoServicePackagePath(name))

	// Generate commands, docker- and makefile
	commands := generate.NewCommandOptions(name)
	generate.Commands(dir, commands)
	generate.Dockerfile(filepath.Join(dir, "Dockerfile"), generate.DockerfileOptions{
		Commands: commands,
	})
	generate.Makefile(filepath.Join(dir, "Makefile"), generate.MakefileOptions{
		Name: name,
	})
}
