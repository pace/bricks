// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/31 by Vincent Landgraf

package generate

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

// CommandOptions are applied when generating the different
// microservice commands
type CommandOptions struct {
	DaemonName  string
	ControlName string
}

// NewCommandOptions generate command names using given name
func NewCommandOptions(name string) CommandOptions {
	return CommandOptions{
		DaemonName:  name + "d",
		ControlName: name + "ctl",
	}
}

// Commands generates the microservice commands based of
// the given path
func Commands(path string, options CommandOptions) {
	// Create directories
	dirs := []string{
		filepath.Join(path, "cmd", options.DaemonName),
		filepath.Join(path, "cmd", options.ControlName),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0770)
		if err != nil {
			log.Fatal(fmt.Printf("Failed to create dir %s: %v", dir, err))
		}
	}

	// Create commands files
	for _, dir := range dirs {
		f, err := os.Create(filepath.Join(dir, "main.go"))
		if err != nil {
			log.Fatal(err)
		}

		templateData := struct {
			Name string
		}{
			Name: filepath.Base(dir),
		}

		err = mainTemplate.Execute(f, templateData)
		if err != nil {
			log.Fatal(err)
		}
	}
}

var mainTemplate = template.Must(template.New("Makefile").
	Parse(`// Copyright © YYYY by PACE Telematics GmbH. All rights reserved.
// Created at YYYY/MM/DD by <NAME OF AUTHOR>

package main

import "fmt"

func main() {
	fmt.Printf("{{ .Name }}")
}
`))
