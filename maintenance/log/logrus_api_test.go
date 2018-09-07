// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"testing"
	"time"
)

func TestLogrusAPI(t *testing.T) {
	Error("Test", 1, time.Now())
	Errorf("Test %d %v", 1, time.Now())

	Warn("Test", 1, time.Now())
	Warnf("Test %d %v", 1, time.Now())

	Info("Test", 1, time.Now())
	Infof("Test %d %v", 1, time.Now())

	Debug("Test", 1, time.Now())
	Debugf("Test %d %v", 1, time.Now())
}
