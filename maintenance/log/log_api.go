// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package log

import (
	"github.com/pace/bricks/maintenance/terminationlog"
)

// Fatal implements log Fatal interface
func Fatal(v ...interface{}) {
	terminationlog.Fatal(v...)
}

// Fatalln implements log Fatalln interface
func Fatalln(v ...interface{}) {
	terminationlog.Fatalln(v...)
}

// Fatalf implements log Fatalf interface
func Fatalf(format string, v ...interface{}) {
	terminationlog.Fatalf(format, v...)
}

// Print implements log Print interface
func Print(v ...interface{}) {
	Debug(v...)
}

// Println implements log Println interface
func Println(v ...interface{}) {
	Debug(v...)
}

// Printf implements log Printf interface
func Printf(format string, v ...interface{}) {
	Debugf(format, v...)
}
