// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"sync"
	"time"
)

const DefaultPollingRate = 1 * time.Second

type msgMeter struct {
	Key   string
	Meter Meter
}

type Poller struct {
	Meters   map[string]Meter
	Handlers []Handler

	PollingRate time.Duration

	initialize sync.Once

	addC    chan msgMeter
	removeC chan string
	handleC chan Handler
}

func (poller *Poller) Init() {
	poller.initialize.Do(poller.init)
}

func (poller *Poller) Add(key string, meter Meter) {
	poller.Init()
	poller.addC <- msgMeter{key, meter}
}

func (poller *Poller) Remove(key string) {
	poller.Init()
	poller.removeC <- key
}

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
			result[Join(prefix, suffix)] = value
		}
	}

	for _, handler := range poller.Handlers {
		handler.HandleMeters(result)
	}
}
