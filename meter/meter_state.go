// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

type State struct {
	value string
	mutex sync.Mutex
}

func (state *State) Change(value string) {
	state.mutex.Lock()

	state.value = value

	state.mutex.Unlock()
}

func (state *State) Reset() {
	state.mutex.Lock()

	state.value = ""

	state.mutex.Unlock()
}

func (state *State) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	state.mutex.Lock()

	if state.value != "" {
		result[state.value] = 1.0
	}

	state.mutex.Unlock()

	return result
}

func GetState(prefix string) *State {
	return GetOrAdd(prefix, new(State)).(*State)
}
