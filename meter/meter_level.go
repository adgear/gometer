// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

// Level reports a recorded value until a new value is recorded. This can be
// useful to record the result of rare or periodic events by ensuring that the
// results are always available and hard to miss.
type Level struct {

	// Value contains the initial value of the level. Should not be read or
	// written after construction.
	Value float64

	mutex sync.Mutex
}

// Record changes the level to the givne value.
func (level *Level) Record(value float64) {
	level.mutex.Lock()

	level.Value = value

	level.mutex.Unlock()
}

// RecordInt is similar to Record but with int values.
func (level *Level) RecordInt(value int) {
	level.Record(float64(value))
}

// ReadMeter returns the currently set level if not equal to 0.
func (level *Level) ReadMeter(_ time.Duration) map[string]float64 {
	result := make(map[string]float64)

	level.mutex.Lock()

	if level.Value != 0 {
		result[""] = level.Value
	}

	level.mutex.Unlock()

	return result
}
