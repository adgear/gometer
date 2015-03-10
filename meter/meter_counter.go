// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync/atomic"
	"time"
)

// Counter counts the number of occurence of an event. Is also completely
// go-routine safe.
type Counter struct{ value uint64 }

// Hit adds 1 to the counter.
func (counter *Counter) Hit() {
	atomic.AddUint64(&counter.value, 1)
}

// Count adds the given value to the counter.
func (counter *Counter) Count(count uint64) {
	atomic.AddUint64(&counter.value, count)
}

// ReadMeter returns the current count and resets the counter's value to 0. The
// returned value is normalized using the given delta to ensure that the value
// always represents a per second value.
func (counter *Counter) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	if value := atomic.SwapUint64(&counter.value, 0); value > 0 {
		result[""] = float64(value) * (float64(time.Second) / float64(delta))
	}

	return result
}

// GetCounter returns the counter registered with the given key or creates a new
// one and registers it.
func GetCounter(prefix string) *Counter {
	return GetOrAdd(prefix, new(Counter)).(*Counter)
}
