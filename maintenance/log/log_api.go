// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// Fatal implements log Fatal interface
func Fatal(v ...interface{}) { log.Fatal().Msg(fmt.Sprint(v...)) }

// Fatalln implements log Fatalln interface
func Fatalln(v ...interface{}) { Fatal(v...) }

// Fatalf implements log Fatalf interface
func Fatalf(format string, v ...interface{}) { log.Fatal().Msg(fmt.Sprintf(format, v...)) }

// Print implements log Print interface
func Print(v ...interface{}) { Debug(v...) }

// Println implements log Println interface
func Println(v ...interface{}) { Debug(v...) }

// Printf implements log Printf interface
func Printf(format string, v ...interface{}) { Debugf(format, v...) }
