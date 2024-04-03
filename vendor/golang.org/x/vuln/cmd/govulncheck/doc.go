// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Govulncheck reports known vulnerabilities that affect Go code. It uses static
analysis of source code or a binary's symbol table to narrow down reports to
only those that could affect the application.

By default, govulncheck makes requests to the Go vulnerability database at
https://vuln.go.dev. Requests to the vulnerability database contain only module
paths, not code or other properties of your program. See
https://vuln.go.dev/privacy.html for more. Use the -db flag to specify a
different database, which must implement the specification at
https://go.dev/security/vuln/database.

Govulncheck looks for vulnerabilities in Go programs using a specific build
configuration. For analyzing source code, that configuration is the Go version
specified by the “go” command found on the PATH. For binaries, the build
configuration is the one used to build the binary. Note that different build
configurations may have different known vulnerabilities.

Govulncheck must be built with Go version 1.18 or later.

# Usage

To analyze source code, run govulncheck from the module directory, using the
same package path syntax that the go command uses:

	$ cd my-module
	$ govulncheck ./...

If no vulnerabilities are found, govulncheck will display a short message. If
there are vulnerabilities, each is displayed briefly, with a summary of a call
stack. The summary shows in brief how the package calls a vulnerable function.
For example, it might say

	main.go:[line]:[column]: mypackage.main calls golang.org/x/text/language.Parse

To control which files are processed, use the -tags flag to provide a
comma-separated list of build tags, and the -test flag to indicate that test
files should be included.

To include more detailed stack traces, pass -show=traces, this will cause it to
print the full call stack for each entry.

To run govulncheck on a compiled binary, pass it the path to the binary file
with the -mode=binary flag:

	$ govulncheck -mode=binary $HOME/go/bin/my-go-program

Govulncheck uses the binary's symbol information to find mentions of vulnerable
functions. Its output omits call stacks, which require source code analysis.

Govulncheck also supports -mode=extract on a Go binary for extraction of minimal
information needed to analyze the binary. This will produce a blob, typically much
smaller than the binary, that can also be passed to govulncheck as an argument with
-mode=binary. The users should not rely on the contents or representation of the blob.

Govulncheck exits successfully (exit code 0) if there are no vulnerabilities,
and exits unsuccessfully if there are. It also exits successfully if the -json flag
is provided, regardless of the number of detected vulnerabilities.

Govulncheck supports streaming JSON. For more details, please see [golang.org/x/vuln/internal/govulncheck].

# Limitations

Govulncheck has these limitations:

  - Govulncheck analyzes function pointer and interface calls conservatively,
    which may result in false positives or inaccurate call stacks in some cases.
  - Calls to functions made using package reflect are not visible to static
    analysis. Vulnerable code reachable only through those calls will not be
    reported. Use of the unsafe package may result in false negatives.
  - Because Go binaries do not contain detailed call information, govulncheck
    cannot show the call graphs for detected vulnerabilities. It may also
    report false positives for code that is in the binary but unreachable.
  - There is no support for silencing vulnerability findings. See https://go.dev/issue/61211 for
    updates.
  - Govulncheck only reads binaries compiled with Go 1.18 and later.
  - For binaries where the symbol information cannot be extracted, govulncheck
    reports vulnerabilities for all modules on which the binary depends.

# Feedback

To share feedback, see https://go.dev/security/vuln#feedback.
*/
package main
