// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// MultiHistogram associates Histogram objects to keys which can be
// selected when recording. Is completely go-routine safe.
type MultiHistogram struct {

	// Size is used to initialize the Size member of the underlying Histogram
	// objects.
	Size int

	// SamplingSeed is used to initialize the SamplingSeed member of underlying
	// Histogram objects.
	SamplingSeed int64

	dists unsafe.Pointer
	mutex sync.Mutex
}

// Record adds the given value to the histogram associated with the given
// key. New keys are lazily created as required.
func (multi *MultiHistogram) Record(key string, value float64) {
	multi.get(key).Record(value)
}

// RecordDuration similar to Record but with time.Duration values.
func (multi *MultiHistogram) RecordDuration(key string, value time.Duration) {
	multi.get(key).RecordDuration(value)
}

// RecordSince records a duration elapsed since the given time for the given
// key.
func (multi *MultiHistogram) RecordSince(key string, t0 time.Time) {
	multi.get(key).RecordSince(t0)
}

// ReadMeter calls ReadMeter on all the underlying histograms where all the
// keys are prefixed by the key name used in the calls to Record.
func (multi *MultiHistogram) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	old := multi.load()
	if old == nil {
		return result
	}

	for prefix, dist := range *old {
		for suffix, value := range dist.ReadMeter(delta) {
			result[Join(prefix, suffix)] = value
		}
	}

	return result
}

func (multi *MultiHistogram) get(key string) *Histogram {
	if dists := multi.load(); dists != nil {
		if dist, ok := (*dists)[key]; ok {
			return dist
		}
	}

	multi.mutex.Lock()
	defer multi.mutex.Unlock()

	oldDists := multi.load()
	if oldDists != nil {
		if dist, ok := (*oldDists)[key]; ok {
			return dist
		}
	}

	newDists := new(map[string]*Histogram)
	*newDists = make(map[string]*Histogram)

	if oldDists != nil {
		for key, dist := range *oldDists {
			(*newDists)[key] = dist
		}
	}

	dist := &Histogram{
		Size:         multi.Size,
		SamplingSeed: multi.SamplingSeed,
	}
	(*newDists)[key] = dist
	multi.store(newDists)

	return dist
}

func (multi *MultiHistogram) load() *map[string]*Histogram {
	return (*map[string]*Histogram)(atomic.LoadPointer(&multi.dists))
}

func (multi *MultiHistogram) store(dists *map[string]*Histogram) {
	atomic.StorePointer(&multi.dists, unsafe.Pointer(dists))
}

// GetMultiHistogram returns the histogram registered with the given key or
// creates a new one and registers it.
func GetMultiHistogram(prefix string) *MultiHistogram {
	return GetOrAdd(prefix, new(MultiHistogram)).(*MultiHistogram)
}
