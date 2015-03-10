// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"testing"
	"time"
)

func TestPoller(t *testing.T) {
	m0 := &Gauge{Value: 1}
	m1 := &Gauge{Value: 2}
	m2 := &Gauge{Value: 3}

	h0 := &TestHandler{T: t}
	h1 := &TestHandler{T: t}

	poller := &Poller{
		Meters:   map[string]Meter{"m0": m0},
		Handlers: []Handler{h0},
	}

	poller.Poll("", 100*time.Millisecond)

	h0.Expect("init", map[string]float64{"m0": 1})

	poller.Add("m1", m1)
	h0.Expect("add-m1", map[string]float64{"m0": 1, "m1": 2})

	poller.Add("m2", m2)
	h0.Expect("add-m2", map[string]float64{"m0": 1, "m1": 2, "m2": 3})

	poller.Remove("m0")
	h0.Expect("rmv-m0", map[string]float64{"m1": 2, "m2": 3})

	poller.Remove("m2")
	h0.Expect("rmv-m2", map[string]float64{"m1": 2})

	poller.Add("m0", m0)
	h0.Expect("add-m0", map[string]float64{"m0": 1, "m1": 2})

	poller.Handle(h1)
	h0.Expect("add-h1", map[string]float64{"m0": 1, "m1": 2})
	h1.Expect("add-h1", map[string]float64{"m0": 1, "m1": 2})
}
