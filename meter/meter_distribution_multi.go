// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type MultiDistribution struct {
	Size         int
	SamplingSeed int64

	dists unsafe.Pointer
	mutex sync.Mutex
}

func (multi *MultiDistribution) Record(key string, value float64) {
	multi.get(key).Record(value)
}

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
