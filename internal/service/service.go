// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/10 by Vincent Landgraf

package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Clone the service into pace path
func Clone(name string) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	SimpleExec("git", "clone", fmt.Sprintf(GitLabTemplate, name), dir)
}

// Path prints the path of the service identified by name to STDOUT
func Path(name string) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
}

// Edit the service with given name in favorite editor, defined
// by env PACE_EDITOR or EDITOR
func Edit(name string) {
	editor, ok := os.LookupEnv("PACE_EDITOR")

	if !ok {
		editor, ok = os.LookupEnv("EDITOR")

		if !ok {
			log.Fatal("No $PACE_EDITOR or $EDITOR defined!")
		}
	}

	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	SimpleExec(editor, dir)
}

// RunOptions fallback cmdName and additional arguments for the run cmd
type RunOptions struct {
	CmdName string   // alternative name for the command of the service
	Args    []string // rest of arguments
}

// Run the service daemon for the given name or use the optional
// cmdname instead
func Run(name string, options RunOptions) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	// identify the go files to run
	var args []string
	if options.CmdName == "" {
		args, err = filepath.Glob(filepath.Join(dir, fmt.Sprintf("cmd/%sd/*.go", name)))
	} else {
		args, err = filepath.Glob(filepath.Join(dir, fmt.Sprintf("cmd/%s/*.go", options.CmdName)))
	}
	if err != nil {
		log.Fatal(err)
	}

	// start go run
	args = append([]string{"run"}, args...)
	args = append(args, options.Args...)
	SimpleExec("go", args...)
}

// TestOptions options to respect when starting a test
type TestOptions struct {
	GoConvey bool
}

// Test execute the gorich or goconvey test runners
func Test(name string, options TestOptions) {
	if options.GoConvey {
		AutoInstall("goconvey", "github.com/smartystreets/goconvey")
	} else {
		AutoInstall("richgo", "github.com/kyoh86/richgo")
	}

	// get dir for the service
	pkg := filepath.Join(GoServicePackagePath(name), "...")

	if options.GoConvey {
		// get dir for the service
		dir, err := GoServicePath(name)
		if err != nil {
			log.Fatal(err)
		}

		SimpleExecInPath(dir, GoBinCommand("goconvey"))
	} else {
		SimpleExec(GoBinCommand("richgo"), "test", "-race", "-cover", pkg)
	}
}

// Lint executes golint or installes if not already installed
func Lint(name string) {
	AutoInstall("golint", "golang.org/x/lint/golint")

	var buf bytes.Buffer
	GoBinCommandText(&buf, "go", "list", filepath.Join(GoServicePackagePath(name), "..."))
	paths := strings.Split(buf.String(), "\n")

	// start go run
	SimpleExec(GoBinCommand("golint"), paths...)
}
