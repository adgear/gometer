// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

type Level struct {
	Value float64
	mutex sync.Mutex
}

func (level *Level) Record(value float64) {
	level.mutex.Lock()

	level.Value = value

	level.mutex.Unlock()
}

func (level *Level) ReadMeter(_ time.Duration) map[string]float64 {
	level.mutex.Lock()
	defer level.mutex.Unlock()

	if level.Value == 0 {
		return map[string]float64{}
	}

	result := map[string]float64{"": level.Value}

	return result
}
