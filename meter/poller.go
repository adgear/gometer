// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

// DefaultPollingRate will be used as the value for PollingRate in Poller if no
// values are provided.
const DefaultPollingRate = 1 * time.Second

// Poller periodically calls ReadMeter on all the meters registered via the Add
// function and forwards the aggregated values to all the configured handlers.
type Poller struct {

	// Meters contains the initial list of meters to be polled. Should not be
	// read or modified after calling Init.
	Meters map[string]Meter

	// Handlers contains the initial list of handlers. Should not be read or
	// modified after calling init.
	Handlers []Handler

	// PollingRate is the frequency at which meters will be polled.
	PollingRate time.Duration

	// KeyPrefix is a prefix applied to all polled keys.
	KeyPrefix string

	initialize sync.Once

	addC    chan msgMeter
	removeC chan string
	handleC chan Handler
}

type msgMeter struct {
	Key   string
	Meter Meter
}

// Init can be optionally used to initialize the poller.
func (poller *Poller) Init() {
	poller.initialize.Do(poller.init)
}

// Add registers the given meter which will be polled periodically and
// associates it with the given key. If the key is already registered then the
// old meter will be replaced by the given meter.
func (poller *Poller) Add(key string, meter Meter) {
	poller.Init()
	poller.addC <- msgMeter{key, meter}
}

// Remove unregisters the given meter which will no longer be polled
// periodically.
func (poller *Poller) Remove(key string) {
	poller.Init()
	poller.removeC <- key
}

// Handle adds the given handler to the list of handler to execute.
func (poller *Poller) Handle(handler Handler) {
	poller.Init()
	poller.handleC <- handler
}

func (poller *Poller) init() {
	if poller.Meters == nil {
		poller.Meters = make(map[string]Meter)
	}

	if poller.PollingRate == 0 {
		poller.PollingRate = DefaultPollingRate
	}

	poller.addC = make(chan msgMeter)
	poller.removeC = make(chan string)
	poller.handleC = make(chan Handler)

	go poller.run()
}

func (poller *Poller) run() {
	tickC := time.Tick(poller.PollingRate)

	for {
		select {
		case msg := <-poller.addC:
			poller.add(msg.Key, msg.Meter)

		case key := <-poller.removeC:
			poller.remove(key)

		case handler := <-poller.handleC:
			poller.handle(handler)

		case <-tickC:
			poller.poll()
		}
	}
}

func (poller *Poller) add(key string, meter Meter) {
	poller.Meters[key] = meter

	// Discard any stale values.
	meter.ReadMeter(poller.PollingRate)
}

func (poller *Poller) remove(key string) {
	delete(poller.Meters, key)
}

func (poller *Poller) handle(handler Handler) {
	poller.Handlers = append(poller.Handlers, handler)
}

func (poller *Poller) poll() {
	result := make(map[string]float64)

	for prefix, meter := range poller.Meters {
		for suffix, value := range meter.ReadMeter(poller.PollingRate) {
			result[Join(poller.KeyPrefix, prefix, suffix)] = value
		}
	}

	for _, handler := range poller.Handlers {
		handler.HandleMeters(result)
	}
}

// DefaultPoller is the Poller object used by the global Add, Remove and Handle
// functions.
var DefaultPoller Poller

// Add associates the given meter with the given key and begins to periodically
// poll the meter.
func Add(key string, meter Meter) {
	DefaultPoller.Add(key, meter)
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
