// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package log

import (
	"github.com/pace/bricks/maintenance/terminationlog"
)

// Fatal implements log Fatal interface.
func Fatal(v ...any) {
	terminationlog.Fatal(v...)
}

// Fatalln implements log Fatalln interface.
func Fatalln(v ...any) {
	terminationlog.Fatalln(v...)
}

// Fatalf implements log Fatalf interface.
func Fatalf(format string, v ...any) {
	terminationlog.Fatalf(format, v...)
}

// Print implements log Print interface.
func Print(v ...any) {
	Debug(v...)
}

// Println implements log Println interface.
func Println(v ...any) {
	Debug(v...)
}

// Printf implements log Printf interface.
func Printf(format string, v ...any) {
	Debugf(format, v...)
}
