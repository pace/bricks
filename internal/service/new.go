// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pace/bricks/internal/service/generate"
)

// NewOptions collection of options to apply while or
// after the creation of the new project.
type NewOptions struct {
	RestSource string // url or path to OpenAPIv3 (json:api) specification
}

// New creates a new directory in the go path.
func New(name string, options NewOptions) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	SimpleExec("git", "init", dir)
	SimpleExecInPath(dir, "git", "remote", "add", "origin", fmt.Sprintf(GitLabTemplate, name))
	log.Printf("Remember to create the %s repository in gitlab: https://your_gitlab_url_goes_here/projects/new\n", name)

	// add REST API if there was a source specified
	if options.RestSource != "" {
		restDir := filepath.Join(dir, "internal", "http", "rest")

		if err := os.MkdirAll(restDir, 0o750); err != nil {
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
		Name:     name,
		Commands: commands,
	})
	generate.Makefile(filepath.Join(dir, "Makefile"), generate.MakefileOptions{
		Name: name,
	})

	SimpleExecInPath(dir, "go", "mod", "vendor")
}
