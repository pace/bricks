// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/10 by Vincent Landgraf
package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// PaceBase for all go projects
const PaceBase = "lab.jamit.de/pace"

// ServiceBase for all go microservices
const ServiceBase = "web/service"

// GitLabTemplate git clone template for cloning repositories
const GitLabTemplate = "git@lab.jamit.de:pace/web/service/%s.git"

// GoPath returns the gopath for the current system,
// uses GOPATH env and fallback to default go dir
func GoPath() string {
	path, ok := os.LookupEnv("GOPATH")
	if !ok {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(usr.HomeDir, "go")
	}

	return path
}

// New creates a new directory in the go path
func New(name string) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	AutoInstallDep()

	SimpleExec("git", "init", dir)
	SimpleExecInPath(dir, "git", "remote", "add", "origin", fmt.Sprintf(GitLabTemplate, name))
	SimpleExecInPath(dir, GoBinCommand("dep"), "init")
	log.Fatalf("Remember to create the %s repository in gitlab: https://lab.jamit.de/projects/new?namespace_id=296\n", name)
}

// Clone the service into gopath
func Clone(name string) {
	// get dir for the service
	dir, err := GoServicePath(name)
	if err != nil {
		log.Fatal(err)
	}

	AutoInstallDep()

	SimpleExec("git", "clone", fmt.Sprintf(GitLabTemplate, name), dir)
	SimpleExecInPath(dir, GoBinCommand("dep"), "ensure", "-v")
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

	AutoInstallDep()

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
	SimpleExecInPath(dir, GoBinCommand("dep"), "ensure")
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

// GoServicePath returns the path of the go service for given name
func GoServicePath(name string) (string, error) {
	return filepath.Abs(filepath.Join(GoPath(), "src", PaceBase, ServiceBase, name))
}

// GoServicePackagePath returns a go package path for given service name
func GoServicePackagePath(name string) string {
	return filepath.Join(PaceBase, ServiceBase, name)
}

// AutoInstallDep installs the go dep tool (will be removed if vgo available)
func AutoInstallDep() {
	AutoInstall("dep", "github.com/golang/dep/cmd/dep")
}

// AutoInstall cmdName if not installed already using go get -u goGetPath
func AutoInstall(cmdName, goGetPath string) {
	if _, err := os.Stat(GoBinCommand(cmdName)); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Installing %s using: go get -u %s\n", cmdName, goGetPath)
		// assume error means no file
		SimpleExec("go", "get", "-u", goGetPath)
	} else if err != nil {
		log.Fatal(err)
	}
}

// GoBinCommand returns the path to a binary installed in the gopath
func GoBinCommand(cmdName string) string {
	return filepath.Join(GoPath(), "bin", cmdName)
}

// SimpleExec executes the command and uses the parent process STDIN,STDOUT,STDERR
func SimpleExec(cmdName string, arguments ...string) {
	cmd := exec.Command(cmdName, arguments...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// SimpleExecInPath executes the command and uses the parent process STDIN,STDOUT,STDERR in passed dir
func SimpleExecInPath(dir, cmdName string, arguments ...string) {
	cmd := exec.Command(cmdName, arguments...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// GoBinCommandText writes the command output to the passed writer
func GoBinCommandText(w io.Writer, cmdName string, arguments ...string) {
	cmd := exec.Command(cmdName, arguments...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
