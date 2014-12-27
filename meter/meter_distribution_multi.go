// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// MultiDistribution associates Distribution objects to keys which can be
// selected when recording. Is completely go-routine safe.
type MultiDistribution struct {

	// Size is used to initialize the Size member of the underlying Distribution
	// objects.
	Size int

	// SamplingSeed is used to initialize the SamplingSeed member of underlying
	// Distribution objects.
	SamplingSeed int64

	dists unsafe.Pointer
	mutex sync.Mutex
}

// Record adds the given value to the distribution associated with the given
// key. New keys are lazily created as required.
func (multi *MultiDistribution) Record(key string, value float64) {
	multi.get(key).Record(value)
}

// RecordDuration similar to Record but with time.Duration values.
func (multi *MultiDistribution) RecordDuration(key string, value time.Duration) {
	multi.get(key).RecordDuration(value)
}

// ReadMeter calls ReadMeter on all the underlying distributions where all the
// keys are prefixed by the key name used in the calls to Record.
func (multi *MultiDistribution) ReadMeter(delta time.Duration) map[string]float64 {
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

func (multi *MultiDistribution) get(key string) *Distribution {
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

	newDists := new(map[string]*Distribution)
	*newDists = make(map[string]*Distribution)

	if oldDists != nil {
		for key, dist := range *oldDists {
			(*newDists)[key] = dist
		}
	}

	dist := &Distribution{
		Size:         multi.Size,
		SamplingSeed: multi.SamplingSeed,
	}
	(*newDists)[key] = dist
	multi.store(newDists)

	return dist
}

func (multi *MultiDistribution) load() *map[string]*Distribution {
	return (*map[string]*Distribution)(atomic.LoadPointer(&multi.dists))
}

func (multi *MultiDistribution) store(dists *map[string]*Distribution) {
	atomic.StorePointer(&multi.dists, unsafe.Pointer(dists))
}
