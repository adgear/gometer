// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type MultiCounter struct {
	counters unsafe.Pointer
	mutex    sync.Mutex
}

func (multi *MultiCounter) RecordHit(key string) {
	multi.get(key).RecordHit()
}

func (multi *MultiCounter) RecordCount(key string, count uint64) {
	multi.get(key).RecordCount(count)
}

func (multi *MultiCounter) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	old := multi.load()
	if old == nil {
		return result
	}

	for prefix, counter := range *old {
		for suffix, value := range counter.ReadMeter(delta) {
			result[Join(prefix, suffix)] = value
		}
	}

	return result
}

func (multi *MultiCounter) get(key string) *Counter {
	if counters := multi.load(); counters != nil {
		if counter, ok := (*counters)[key]; ok {
			return counter
		}
	}

	multi.mutex.Lock()
	defer multi.mutex.Unlock()

	oldCounters := multi.load()
	if oldCounters != nil {
		if counter, ok := (*oldCounters)[key]; ok {
			return counter
		}
	}

	newCounters := new(map[string]*Counter)
	*newCounters = make(map[string]*Counter)

	if oldCounters != nil {
		for key, counter := range *oldCounters {
			(*newCounters)[key] = counter
		}
	}

	counter := new(Counter)
	(*newCounters)[key] = counter
	multi.store(newCounters)

	return counter
}

func (multi *MultiCounter) load() *map[string]*Counter {
	return (*map[string]*Counter)(atomic.LoadPointer(&multi.counters))
}

func (multi *MultiCounter) store(counters *map[string]*Counter) {
	atomic.StorePointer(&multi.counters, unsafe.Pointer(counters))
}
