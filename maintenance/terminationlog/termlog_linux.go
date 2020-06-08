// Package terminationlog helps to fill the kubernetes termination log.
// From the doc:
// Termination messages provide a way for containers to write information
// about fatal events to a location where it can be easily retrieved and
// surfaced by tools like dashboards and monitoring software. In most
// cases, information that you put in a termination message should also
// be written to the general Kubernetes logs.
package terminationlog

import (
	"os"
	"syscall"
)

// termLog default location of kubernetes termination log
const termLog = "/dev/termination-log"

func init() {
	file, err := os.OpenFile(termLog, O_RDWR, 0666)

	if err == nil {
		logFile = file

		// redirect stderr to the termLog
		syscall.Dup2(int(logFile.Fd()), 2) // nolint: errcheck
	}
}
