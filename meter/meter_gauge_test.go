// Copyright (c) 2014 Datacratic. All rights reserved.
//
// This meter is dirt simple so we won't bother with perf or parallel tests.

package meter

import (
	"testing"
)

func TestGauge(t *testing.T) {
	var gauge Gauge

	gauge.Change(1)
	CheckGauge(t, &gauge, 1)

	gauge.Change(2)
	gauge.Change(3)
	CheckGauge(t, &gauge, 3)

	CheckGauge(t, &gauge, 3)
}

func CheckGauge(t *testing.T, gauge *Gauge, exp float64) {
	if value := gauge.ReadMeter(0)[""]; value != exp {
		t.Errorf("FAIL: value=%f != %f", value, exp)
	}
}
