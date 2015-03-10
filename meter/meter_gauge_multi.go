// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// MultiGauge associates Gauge objects to keys which can be selected when
// recording. Is completely go-routine safe.
type MultiGauge struct {
	gauges unsafe.Pointer
	mutex  sync.Mutex
}

// Change records the given value with the gauge associated with the given
// key. New Keys are lazily created as required.
func (multi *MultiGauge) Change(key string, value float64) {
	multi.get(key).Change(value)
}

// ChangeDuration similar to Change but with a time.Duration value.
func (multi *MultiGauge) ChangeDuration(key string, duration time.Duration) {
	multi.get(key).ChangeDuration(duration)
}

// ChangeSince records a duration elapsed since the given time with the given
// key.
func (multi *MultiGauge) ChangeSince(key string, t0 time.Time) {
	multi.get(key).ChangeSince(t0)
}

// ReadMeter calls ReadMeter on all underlying gauges where all the keys are
// prefixed by the key name used in the calls to Record.
func (multi *MultiGauge) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	old := multi.load()
	if old == nil {
		return result
	}

	for prefix, gauge := range *old {
		for suffix, value := range gauge.ReadMeter(delta) {
			result[Join(prefix, suffix)] = value
		}
	}

	return result
}

func (multi *MultiGauge) get(key string) *Gauge {
	if gauges := multi.load(); gauges != nil {
		if gauge, ok := (*gauges)[key]; ok {
			return gauge
		}
	}

	multi.mutex.Lock()
	defer multi.mutex.Unlock()

	oldGauges := multi.load()
	if oldGauges != nil {
		if gauge, ok := (*oldGauges)[key]; ok {
			return gauge
		}
	}

	newGauges := new(map[string]*Gauge)
	*newGauges = make(map[string]*Gauge)

	if oldGauges != nil {
		for key, gauge := range *oldGauges {
			(*newGauges)[key] = gauge
		}
	}

	gauge := new(Gauge)
	(*newGauges)[key] = gauge
	multi.store(newGauges)

	return gauge
}

func (multi *MultiGauge) load() *map[string]*Gauge {
	return (*map[string]*Gauge)(atomic.LoadPointer(&multi.gauges))
}

func (multi *MultiGauge) store(gauges *map[string]*Gauge) {
	atomic.StorePointer(&multi.gauges, unsafe.Pointer(gauges))
}

// GetMultiGauge returns the gauge registered with the given key or creates a
// new one and registers it.
func GetMultiGauge(prefix string) *MultiGauge {
	return GetOrAdd(prefix, new(MultiGauge)).(*MultiGauge)
}
