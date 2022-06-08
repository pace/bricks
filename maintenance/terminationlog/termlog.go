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
	"github.com/rs/zerolog/log"
	"os"
)

var logFile *os.File

const LogFileLimit = 4096 // bytes;

// Fatalf implements log Fatalf interface
func Fatalf(format string, v ...interface{}) {
	if logFile != nil {
		fmt.Fprint(logFile, limitLogFileOutput(fmt.Sprintf(format, v...)))
	}

	log.Fatal().Msg(fmt.Sprintf(format, v...))
}

// Fatal implements log Fatal interface
func Fatal(v ...interface{}) {
	if logFile != nil {
		fmt.Fprint(logFile, limitLogFileOutput(fmt.Sprint(v...)))
	}

	log.Fatal().Msg(fmt.Sprint(v...))
}

// Fatalln implements log Fatalln interface
func Fatalln(v ...interface{}) {
	Fatal(v...)
}

func limitLogFileOutput(s string) string {
	sb := []byte(s)
	limit := len(sb)
	if limit > LogFileLimit {
		limit = LogFileLimit
	}

	return string(sb[:limit])
}
