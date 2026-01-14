// Copyright © 2018–2026 by PACE Mobility GmbH. All rights reserved.

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
