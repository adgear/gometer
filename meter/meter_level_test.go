// Copyright (c) 2014 Datacratic. All rights reserved.
//
// This meter is dirt simple so we won't bother with perf or parallel tests.

package meter

import (
	"testing"
)

func TestLevel(t *testing.T) {
	var level Level

	level.Record(1)
	CheckLevel(t, &level, 1)

	level.Record(2)
	level.Record(3)
	CheckLevel(t, &level, 3)

	CheckLevel(t, &level, 3)
}

func CheckLevel(t *testing.T, level *Level, exp float64) {
	if value := level.ReadMeter(0)[""]; value != exp {
		t.Errorf("FAIL: value=%f != %f", value, exp)
	}
}
