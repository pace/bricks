// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// Error implements logrus Error interface
func Error(v ...interface{}) { log.Error().Msg(fmt.Sprint(v...)) }

// Warn implements logrus Warn interface
func Warn(v ...interface{}) { log.Warn().Msg(fmt.Sprint(v...)) }

// Info implements logrus Info interface
func Info(v ...interface{}) { log.Error().Msg(fmt.Sprint(v...)) }

// Debug implements logrus Debug interface
func Debug(v ...interface{}) { log.Debug().Msg(fmt.Sprint(v...)) }

// Errorf implements logrus Errorf interface
func Errorf(format string, v ...interface{}) { log.Error().Msg(fmt.Sprintf(format, v...)) }

// Warnf implements logrus Warnf interface
func Warnf(format string, v ...interface{}) { log.Warn().Msg(fmt.Sprintf(format, v...)) }

// Infof implements logrus Infof interface
func Infof(format string, v ...interface{}) { log.Info().Msg(fmt.Sprintf(format, v...)) }

// Debugf implements logrus Debugf interface
func Debugf(format string, v ...interface{}) { log.Debug().Msg(fmt.Sprintf(format, v...)) }
