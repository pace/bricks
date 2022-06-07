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
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
)

var logFile *os.File

const StackTraceLimit = 100

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// Fatalf implements log Fatalf interface
func Fatalf(format string, v ...interface{}) {
	if logFile != nil {
		if fs, err := extractErrorFrames(v); err == nil {
			fmt.Fprintf(logFile, format, fs)
		} else {
			fmt.Fprintf(logFile, format, v...)
		}
	}

	log.Fatal().Msg(fmt.Sprintf(format, v...))
}

// Fatal implements log Fatal interface
func Fatal(v ...interface{}) {
	if logFile != nil {
		if fs, err := extractErrorFrames(v); err == nil {
			fmt.Fprint(logFile, fs)
		} else {
			fmt.Fprint(logFile, v...)
		}
	}

	log.Fatal().Msg(fmt.Sprint(v...))
}

// Fatalln implements log Fatalln interface
func Fatalln(v ...interface{}) {
	Fatal(v...)
}

func extractErrorFrames(v ...interface{}) (string, error) {
	if len(v) != 1 {
		return "", fmt.Errorf("value contains multiple elements")
	}

	err, ok := v[0].(error)
	if !ok {
		return "", fmt.Errorf("value element is not an error")
	}

	if str, ok := err.(stackTracer); ok {
		st := str.StackTrace()
		if len(st) > StackTraceLimit {
			return fmt.Sprintf("%+v", st[0:StackTraceLimit]), nil
		} else {
			return fmt.Sprintf("%+v", st[0:]), nil
		}
	} else {
		return "", fmt.Errorf("error does not implement stackTracer")
	}
}
