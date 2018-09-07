// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"testing"
	"time"
)

func TestLogAPI(t *testing.T) {
	Print("Test", 1, time.Now())
	Println("Test", 1, time.Now())
	Printf("Test %d %v", 1, time.Now())
}
