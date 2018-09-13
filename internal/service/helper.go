// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/31 by Vincent Landgraf

package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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

// PacePath returns the pace path for the current system,
// uses PACE_PATH env and fallback to default go dir
func PacePath() string {
	path, ok := os.LookupEnv("PACE_PATH")
	if !ok {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(usr.HomeDir, "PACE")
	}

	return path
}

// GoServicePath returns the path of the go service for given name
func GoServicePath(name string) (string, error) {
	return filepath.Abs(filepath.Join(PacePath(), ServiceBase, name))
}

// GoServicePackagePath returns a go package path for given service name
func GoServicePackagePath(name string) string {
	return filepath.Join(PaceBase, ServiceBase, name)
}

// AutoInstall cmdName if not installed already using go get -u goGetPath
func AutoInstall(cmdName, goGetPath string) {
	if _, err := os.Stat(GoBinCommand(cmdName)); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Installing %s using: go get -u %s\n", cmdName, goGetPath) // nolint: errcheck
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
	cmd := exec.Command(cmdName, arguments...) // nolint: gosec
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
	cmd := exec.Command(cmdName, arguments...) // nolint: gosec
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
	cmd := exec.Command(cmdName, arguments...) // nolint: gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
