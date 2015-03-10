// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

// Poller periodically calls ReadMeter on all the meters registered via the Add
// function and forwards the aggregated values to all the configured handlers.
type Poller struct {

	// Meters contains the initial list of meters to be polled. Should not be
	// read or modified after calling Init.
	Meters map[string]Meter

	// Handlers contains the initial list of handlers. Should not be read or
	// modified after calling init.
	Handlers []Handler

	mutex sync.Mutex

	rate   time.Duration
	prefix string
}

// Get returns the meter associated with the given key or nil if no such meter
// exists.
func (poller *Poller) Get(key string) Meter {
	poller.mutex.Lock()
	defer poller.mutex.Unlock()

	return poller.Meters[key]
}

// Add registers the given meter which will be polled periodically and
// associates it with the given key. If the key is already registered then the
// old meter will be replaced by the given meter.
func (poller *Poller) Add(key string, meter Meter) bool {
	poller.mutex.Lock()
	defer poller.mutex.Unlock()

	meter.ReadMeter(poller.rate)

	if poller.Meters == nil {
		poller.Meters = make(map[string]Meter)
	}

	if _, ok := poller.Meters[key]; ok {
		return false
	}

	poller.Meters[key] = meter
	return true
}

// Remove unregisters the given meter which will no longer be polled
// periodically.
func (poller *Poller) Remove(key string) {
	poller.mutex.Lock()
	defer poller.mutex.Unlock()

	if poller.Meters != nil {
		delete(poller.Meters, key)
	}
}

// Handle adds the given handler to the list of handler to execute.
func (poller *Poller) Handle(handler Handler) {
	poller.mutex.Lock()
	defer poller.mutex.Unlock()

	poller.Handlers = append(poller.Handlers, handler)
}

// Poll starts a background goroutine which will periodically poll the
// registered meters at the given rate and prepend all the polled keys with the
// given prefix.
func (poller *Poller) Poll(prefix string, rate time.Duration) {
	poller.mutex.Lock()
	defer poller.mutex.Unlock()

	if poller.rate != 0 {
		panic("poller already started")
	}
	poller.rate = rate
	poller.prefix = prefix

	go func() {
		for tickC := time.Tick(rate); ; <-tickC {
			poller.poll()
		}
	}()
}

func (poller *Poller) poll() {
	poller.mutex.Lock()
	defer poller.mutex.Unlock()

	result := make(map[string]float64)

	for prefix, meter := range poller.Meters {
		for suffix, value := range meter.ReadMeter(poller.rate) {
			result[Join(poller.prefix, prefix, suffix)] = value
		}
	}

	for _, handler := range poller.Handlers {
		handler.HandleMeters(result)
	}
}

// DefaultPoller is the Poller object used by the global Add, Remove and Handle
// functions.
var DefaultPoller Poller

// Get returns the meter associated with the given key or nil if no such meter
// exists.
func Get(key string) Meter {
	return DefaultPoller.Get(key)
}

// Add associates the given meter with the given key and begins to periodically
// poll the meter.
func Add(key string, meter Meter) bool {
	return DefaultPoller.Add(key, meter)
}

// GetOrAdd registers the given meter with the given key if it's not already
// registered or returns the meter associated with the given key.
func GetOrAdd(key string, meter Meter) Meter {
	for {
		if old := Get(key); old != nil {
			return old
		}

		if Add(key, meter) {
			return meter
		}
	}
}

// Remove removes any meters associated with the key which will no longer be
// polled.
func Remove(key string) {
	DefaultPoller.Remove(key)
}

// Handle adds the given handler to the list of handlers to be executed after
// polling the meters.
func Handle(handler Handler) {
	DefaultPoller.Handle(handler)
}

// Poll starts a background goroutine which will periodically poll the
// registered meters at the given rate and prepend all the polled keys with the
// given prefix.
func Poll(prefix string, rate time.Duration) {
	DefaultPoller.Poll(prefix, rate)
}
