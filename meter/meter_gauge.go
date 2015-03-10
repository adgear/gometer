// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

// Gauge reports a recorded value until a new value is recorded. This can be
// useful to record the result of rare or periodic events by ensuring that the
// results are always available and hard to miss.
type Gauge struct {

	// Value contains the initial value of the gauge. Should not be read or
	// written after construction.
	Value float64

	mutex sync.Mutex
}

// Change changes the recorded value.
func (gauge *Gauge) Change(value float64) {
	gauge.mutex.Lock()

	gauge.Value = value

	gauge.mutex.Unlock()
}

// ChangeDuration similar to Change but with a time.Duration value.
func (gauge *Gauge) ChangeDuration(duration time.Duration) {
	gauge.Change(float64(duration) / float64(time.Second))
}

// ChangeSince records a duration elapsed since the given time.
func (gauge *Gauge) ChangeSince(t0 time.Time) {
	gauge.ChangeDuration(time.Since(t0))
}

// ReadMeter returns the currently set value if not equal to 0.
func (gauge *Gauge) ReadMeter(_ time.Duration) map[string]float64 {
	result := make(map[string]float64)

	gauge.mutex.Lock()

	if gauge.Value != 0 {
		result[""] = gauge.Value
	}

	gauge.mutex.Unlock()

	return result
}

// GetGauge returns the gauge registered with the given key or creates a new one
// and registers.
func GetGauge(prefix string) *Gauge {
	return GetOrAdd(prefix, new(Gauge)).(*Gauge)
}
