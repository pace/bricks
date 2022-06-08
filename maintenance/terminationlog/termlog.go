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

const LogFileLimit = 4096 // bytes

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// Fatalf implements log Fatalf interface
func Fatalf(format string, v ...interface{}) {
	if logFile != nil {
		fmt.Fprint(logFile, buildTerminationLogOutputf(format, v...))
	}

	log.Fatal().Msg(fmt.Sprintf(format, v...))
}

// Fatal implements log Fatal interface
func Fatal(v ...interface{}) {
	if logFile != nil {
		fmt.Fprint(logFile, buildTerminationLogOutput(v...))
	}

	log.Fatal().Msg(fmt.Sprint(v...))
}

// Fatalln implements log Fatalln interface
func Fatalln(v ...interface{}) {
	Fatal(v...)
}

func buildTerminationLogOutputf(f string, v ...interface{}) string {
	if res, err := extractErrorFrames(v...); err == nil {
		vs := make([]interface{}, 0)
		vs = append(vs, f)
		vs = append(vs, res...)
		return buildOutput(vs...)
	}

	return buildOutput(fmt.Sprintf(f, v...))
}

func buildTerminationLogOutput(v ...interface{}) string {
	if res, err := extractErrorFrames(v...); err == nil {
		return buildOutput(res...)
	}

	return buildOutput(v...)
}

func buildOutput(v ...interface{}) string {
	sb := make([]byte, 0)
	for _, f := range v {
		s := fmt.Sprintf("%+v\n", f)
		b := []byte(s)
		if len(b) <= LogFileLimit && len(sb)+len(b) <= LogFileLimit {
			sb = append(sb, b...)
		} else {
			break
		}
	}
	return string(sb)
}

func extractErrorFrames(v ...interface{}) ([]interface{}, error) {
	if len(v) != 1 {
		return nil, fmt.Errorf("value contains multiple elements")
	}

	err, ok := v[0].(error)
	if !ok {
		return nil, fmt.Errorf("value element is not an error")
	}

	if str, ok := err.(stackTracer); ok {
		st := str.StackTrace()
		vs := make([]interface{}, 0)
		vs = append(vs, fmt.Sprintf("%s", err.Error()))
		for _, f := range st {
			vs = append(vs, f)
		}
		return vs, nil
	} else {
		return nil, fmt.Errorf("error does not implement stackTracer")
	}
}
