// Package terminationlog helps to fill the kubernetes termination log.
// From the doc:
// Termination messages provide a way for containers to write information
// about fatal events to a location where it can be easily retrieved and
// surfaced by tools like dashboards and monitoring software. In most
// cases, information that you put in a termination message should also
// be written to the general Kubernetes logs.
package terminationlog

import (
	"fmt"
	"os"
	"syscall"

	"github.com/rs/zerolog/log"
)

// termLog default location of kubernetes termination log
const termLog = "/dev/termination-log"

var logFile *os.File

func init() {
	file, err := os.Create(termLog)

	if err == nil {
		logFile = file

		// redirect stderr to the termLog
		syscall.Dup2(int(logFile.Fd()), 2)
	}
}

// Fatalf implements log Fatalf interface
func Fatalf(format string, v ...interface{}) {
	if logFile != nil {
		fmt.Fprintf(logFile, format, v...)
	}

	log.Fatal().Msg(fmt.Sprintf(format, v...))
}

// Fatal implements log Fatal interface
func Fatal(v ...interface{}) {
	if logFile != nil {
		fmt.Fprint(logFile, v...)
	}

	log.Fatal().Msg(fmt.Sprint(v...))
}

// Fatalln implements log Fatalln interface
func Fatalln(v ...interface{}) {
	Fatal(v...)
}
