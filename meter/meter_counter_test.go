// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"testing"
	"time"
)

func TestCounter_Normalize(t *testing.T) {
	var counter Counter

	ExpectCount(t, &counter, time.Second, 100)
	ExpectCount(t, &counter, 2*time.Second, 50)
	ExpectCount(t, &counter, 500*time.Millisecond, 200)
}

func ExpectCount(t *testing.T, counter *Counter, delta time.Duration, exp float64) {
	counter.Count(100)
	if value := counter.ReadMeter(delta)[""]; value != exp {
		t.Errorf("FAIL: delta=%s -> value=%f != exp=%f", delta, value, exp)
	}
}
