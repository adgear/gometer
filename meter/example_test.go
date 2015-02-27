// Copyright (c) 2014 Datacratic. All rights reserved.

package meter_test

import (
	"github.com/datacratic/gometer/meter"

	"fmt"
	"sort"
)

func ExampleMeter() {

	// First we need to define our meters where all the recorded stats will be
	// stored. Meters comes in 3 flavours (Counter, Histogram and Gauge)
	// along with their multi variant (MultiCounter, MultiHistogram and
	// MultiGauge).
	var counter meter.Counter
	var dist meter.Histogram
	var gauge meter.Gauge
	var multi meter.MultiCounter

	// Next we register our meters with the global meter poller which will
	// periodically read the values of the counters. Note that meters can also
	// be unregistered via the Remove function.
	meter.Add("meters.hits", &counter)
	meter.Add("meters.latency", &dist)
	meter.Add("meters.gauge", &gauge)
	meter.Add("meters.result", &multi)

	// We then need something that will listen to the values aggregated by the
	// poller. These handlers must first be registered to a poller via the
	// Handle function.
	//
	//Meter comes with several of these (eg. CarbonHandler, HTTPHandler, etc.)
	//but for this example we'll create a simple handler that logs to a channel.
	resultC := make(chan map[string]float64)
	handler := func(values map[string]float64) { resultC <- values }
	meter.Handle(meter.HandlerFunc(handler))

	// Next up we'll record some values with our meters.

	counter.Hit()
	counter.Count(10)

	for i := 0; i < 100; i++ {
		dist.Record(float64(i))
	}

	gauge.Change(5)

	multi.Hit("err")
	multi.Count("ok", 10)

	// Finally, we'll finish off the test by reading the value and printing them
	// out.
	SortAndPrint(<-resultC)

	// Output:
	// meters.gauge: 5.000000
	// meters.hits: 11.000000
	// meters.latency.count: 100.000000
	// meters.latency.max: 99.000000
	// meters.latency.min: 0.000000
	// meters.latency.p50: 50.000000
	// meters.latency.p90: 90.000000
	// meters.latency.p99: 99.000000
	// meters.result.err: 1.000000
	// meters.result.ok: 10.000000
}

// SortAndPrint prints the map in a deterministic manner such that we can
// reliably check the output of our example. This is strictly boilerplate for
// the purpose of the example and is not required in actual code.
func SortAndPrint(values map[string]float64) {
	var keys []string

	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("%s: %f\n", key, values[key])
	}

}
