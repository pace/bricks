// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scan

import (
	"errors"
	"strings"
)

//lint:file-ignore ST1005 Ignore staticcheck message about error formatting
var (
	// ErrVulnerabilitiesFound indicates that vulnerabilities were detected
	// when running govulncheck. This returns exit status 3 when running
	// without the -json flag.
	errVulnerabilitiesFound = &exitCodeError{message: "vulnerabilities found", code: 3}

	// errHelp indicates that usage help was requested.
	errHelp = &exitCodeError{message: "help requested", code: 0}

	// errUsage indicates that there was a usage error on the command line.
	//
	// In this case, we assume that the user does not know how to run
	// govulncheck, and print the usage message with exit status 2.
	errUsage = &exitCodeError{message: "invalid usage", code: 2}

	// errGoVersionMismatch is used to indicate that there is a mismatch between
	// the Go version used to build govulncheck and the one currently on PATH.
	errGoVersionMismatch = errors.New(`Loading packages failed, possibly due to a mismatch between the Go version
used to build govulncheck and the Go version on PATH. Consider rebuilding
govulncheck with the current Go version.`)

	// errNoGoMod indicates that a go.mod file was not found in this module.
	errNoGoMod = errors.New(`no go.mod file

govulncheck only works with Go modules. Try navigating to your module directory.
Otherwise, run go mod init to make your project a module.

See https://go.dev/doc/modules/managing-dependencies for more information.`)

	// errNoBinaryFlag indicates that govulncheck was run on a file, without
	// the -mode=binary flag.
	errNoBinaryFlag = errors.New(`By default, govulncheck runs source analysis on Go modules.

Did you mean to run govulncheck with -mode=binary?

For details, run govulncheck -h.`)
)

type exitCodeError struct {
	message string
	code    int
}

func (e *exitCodeError) Error() string { return e.message }
func (e *exitCodeError) ExitCode() int { return e.code }

// isGoVersionMismatchError checks if err is due to mismatch between
// the Go version used to build govulncheck and the one currently
// on PATH.
func isGoVersionMismatchError(err error) bool {
	msg := err.Error()
	// See golang.org/x/tools/go/packages/packages.go.
	return strings.Contains(msg, "This application uses version go") &&
		strings.Contains(msg, "It may fail to process source files")
}
