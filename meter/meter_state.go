// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

// State reports the value 1.0 for a given state until a new state is
// recorded. This can be useful to record infrequent changes in program state.
type State struct {
	value string
	mutex sync.Mutex
}

// Change changes the recorded state. An empty string will not output any
// values.
func (state *State) Change(value string) {
	state.mutex.Lock()

	state.value = value

	state.mutex.Unlock()
}

// Reset disables any values outputed by this state.
func (state *State) Reset() {
	state.Change("")
}

// ReadMeter returns the currently set state with a value of 1.0 or nothing if
// the state is an empty string.
func (state *State) ReadMeter(delta time.Duration) map[string]float64 {
	result := make(map[string]float64)

	state.mutex.Lock()

	if state.value != "" {
		result[state.value] = 1.0
	}

	state.mutex.Unlock()

	return result
}

// GetState returns the state registered with the given key or creates a new one
// and registers it.
func GetState(prefix string) *State {
	return GetOrAdd(prefix, new(State)).(*State)
}
