// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type MultiLevel struct {
	levels unsafe.Pointer
	mutex  sync.Mutex
}

func (multi *MultiLevel) Record(key string, value float64) {
	multi.get(key).Record(value)
}

func (multi *MultiLevel) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	old := multi.load()
	if old == nil {
		return result
	}

	for prefix, level := range *old {
		for suffix, value := range level.ReadMeter(delta) {
			result[Join(prefix, suffix)] = value
		}
	}

	return result
}

func (multi *MultiLevel) get(key string) *Level {
	if levels := multi.load(); levels != nil {
		if level, ok := (*levels)[key]; ok {
			return level
		}
	}

	multi.mutex.Lock()
	defer multi.mutex.Unlock()

	oldLevels := multi.load()
	if oldLevels != nil {
		if level, ok := (*oldLevels)[key]; ok {
			return level
		}
	}

	newLevels := new(map[string]*Level)
	*newLevels = make(map[string]*Level)

	if oldLevels != nil {
		for key, level := range *oldLevels {
			(*newLevels)[key] = level
		}
	}

	level := new(Level)
	(*newLevels)[key] = level
	multi.store(newLevels)

	return level
}

func (multi *MultiLevel) load() *map[string]*Level {
	return (*map[string]*Level)(atomic.LoadPointer(&multi.levels))
}

func (multi *MultiLevel) store(levels *map[string]*Level) {
	atomic.StorePointer(&multi.levels, unsafe.Pointer(levels))
}
